package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/headblockhead/railreader/darwin"
	darwinconn "github.com/headblockhead/railreader/darwin/connection"
	darwindb "github.com/headblockhead/railreader/darwin/database"
	"github.com/jackc/pgx/v5"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
)

type ServeCommand struct {
	Darwin struct {
		// TODO: secrets management tool?
		Kafka struct {
			Host              string        `group:"Darwin Push Port client:" env:"DARWIN_KAFKA_HOST" required:"" help:"Kafka server hostname and port."`
			GroupID           string        `group:"Darwin Push Port client:" env:"DARWIN_KAFKA_GROUP" required:"" name:"group"`
			Topic             string        `group:"Darwin Push Port client:" env:"DARWIN_KAFKA_TOPIC" required:""`
			Username          string        `group:"Darwin Push Port client:" env:"DARWIN_KAFKA_USERNAME" required:""`
			Password          string        `group:"Darwin Push Port client:" env:"DARWIN_KAFKA_PASSWORD" required:""`
			ConnectionTimeout time.Duration `group:"Darwin Push Port client:" default:"10s"`
		} `embed:"" prefix:"kafka."`

		Database struct {
			URL string `group:"Darwin Push Port client:" env:"DARWIN_POSTGRESQL_URL" required:"" help:"PostgreSQL database URL to store Darwin data in."`
		} `embed:"" prefix:"database."`

		QueueSize int `group:"Darwin Push Port client:" default:"32" help:"Maximum number of incoming messages to queue for processing at once. This does not affect data integrity, but will affect memory usage, bandwidth usage on startup, and how long it will take for the server to cleanly exit."`
	} `embed:"" prefix:"darwin."`

	Socket struct {
		Enable bool `group:"Socket:" default:"true" help:"Enable local control of the database via an unauthenticated socket. Access can be limited using file permissions."`
		// TODO: grab location from SYSTEMD.
		Location string `group:"Socket:" help:"Path of the socket file to create."`
		Mode     string `group:"Socket:" default:"600" help:"File mode for the socket. The socket file is owned by the user running the program."`
	} `embed:"" prefix:"socket."`

	/* HTTP struct {*/
	/*Host string `group:"HTTP Server:" env:"HTTP_HOST" default:":8080" help:"HTTP server hostname and port."`*/
	/*} `embed:"" prefix:"http."`*/

	Logging struct {
		Level string `enum:"debug,info,warn,error" default:"warn"`
		Type  string `enum:"json,console" default:"json"`
	} `embed:"" prefix:"logging."`
}

type kafkaConnection interface {
	FetchMessage(ctx context.Context) (kafka.Message, error)
	CommitMessage(msg kafka.Message) error
	Close() error
}

type postgreSQLDatabase interface {
	BeginTx() (pgx.Tx, error)
	Close(timeout time.Duration) error
}

type messageHandler interface {
	Handle(msg kafka.Message) error
}

func (c ServeCommand) Run() error {
	log := getLogger(c.Logging.Level, c.Logging.Type == "json")

	messageFetcherContext, messageFetcherCancel := context.WithCancel(context.Background())

	go func() {
		signalchan := make(chan os.Signal, 1)
		defer close(signalchan)
		signal.Notify(signalchan, syscall.SIGINT, syscall.SIGTERM)

		alreadyTerminating := false
		for {
			signal := <-signalchan // block until a signal is received
			if alreadyTerminating {
				log.Error("received multiple exit signals, exiting immediately")
				os.Exit(130)
			}
			alreadyTerminating = true
			log.Warn(signal.String() + " received, stopping gracefully...")
			messageFetcherCancel()
		}
	}()

	databaseContext := context.Background()
	darwinDatabase, err := darwindb.New(databaseContext, log.With(slog.String("source", "darwin.database")), c.Darwin.Database.URL)
	if err != nil {
		return fmt.Errorf("error connecting to darwin database: %w", err)
	}

	kafkaContext := context.Background()
	darwinKafkaConnection := darwinconn.New(kafkaContext, log.With(slog.String("source", "darwin.connection")), kafka.ReaderConfig{
		Brokers: []string{c.Darwin.Kafka.Host},
		GroupID: c.Darwin.Kafka.GroupID,
		Topic:   c.Darwin.Kafka.Topic,
		Dialer: &kafka.Dialer{
			Timeout:   c.Darwin.Kafka.ConnectionTimeout,
			DualStack: true,
			SASLMechanism: plain.Mechanism{
				Username: c.Darwin.Kafka.Username,
				Password: c.Darwin.Kafka.Password,
			},
			TLS: &tls.Config{},
		},
	})
	darwinKafkaMessages := make(chan kafka.Message, c.Darwin.QueueSize)

	messageHandlerContext := context.Background()
	darwinMessageHandler := darwin.NewMessageHandler(messageHandlerContext, log.With(slog.String("source", "darwin.handler")), darwinDatabase)

	var fetcherGroup sync.WaitGroup
	fetcherGroup.Go(func() {
		fetchMessages(messageFetcherContext, log.With(slog.String("source", "darwin.fetcher")), darwinKafkaMessages, darwinKafkaConnection)
		close(darwinKafkaMessages)
	})
	var processorGroup sync.WaitGroup
	processorGroup.Go(func() {
		processMessages(log.With(slog.String("source", "darwin.processor")), darwinKafkaMessages, darwinKafkaConnection, darwinMessageHandler)
	})

	// The fetcher group will run until messageFetcherContext is cancelled (when the program receives an exit signal).
	fetcherGroup.Wait()
	log.Info("waiting to finish processing all queued messages")
	processorGroup.Wait()

	log.Info("closing connections")
	if err := darwinKafkaConnection.Close(); err != nil {
		log.Error("error closing darwin kafka connection", slog.Any("error", err))
	}
	if err := darwinDatabase.Close(5 * time.Second); err != nil {
		log.Error("error closing darwin database connection", slog.Any("error", err))
	}
	return nil
}

// fetchMessages will run until the context is cancelled.
func fetchMessages(ctx context.Context, log *slog.Logger, messages chan<- kafka.Message, connection kafkaConnection) {
	log.Debug("starting message fetcher")
	for {
		message, err := connection.FetchMessage(ctx)
		if err != nil {
			if err == context.Canceled {
				break
			}
			log.Warn("error fetching message", slog.Any("error", err))
			continue
		}
		messages <- message
	}
	log.Debug("stopped fetching new messages")
}

// processMessages will run until there are no more messages to process (the channel is closed and there are 0 messages remaining in it).
func processMessages(log *slog.Logger, messages <-chan kafka.Message, connection kafkaConnection, messageHandler messageHandler) {
	log.Debug("starting message processor")
	for msg := range messages {
		if err := messageHandler.Handle(msg); err != nil {
			log.Error("error handling message", slog.Any("error", err))
			continue
		}
		if err := connection.CommitMessage(msg); err != nil {
			log.Error("error committing message", slog.Any("error", err))
			continue
		}
	}
	log.Debug("all queued messages processed")
}
