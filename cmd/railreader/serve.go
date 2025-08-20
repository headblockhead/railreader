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

	darwinconnection "github.com/headblockhead/railreader/darwin/connection"
	darwindb "github.com/headblockhead/railreader/darwin/db"
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
		} `embed:"" prefix:"db."`

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

type postgresqlConnection interface {
	BeginTx() (pgx.Tx, error)
	Close(timeout time.Duration) error
}

type message interface {
	Execute(ctx context.Context, tx pgx.Tx) error
}

func (c ServeCommand) Run() error {
	log := getLogger(c.Logging.Level, c.Logging.Type == "json")

	// When the message fetcher context is cancelled, the program will stop fetching new messages.
	messageFetcherContext, messageFetcherCancel := context.WithCancel(context.Background())
	kafkaContext := context.Background()
	databaseContext := context.Background()

	go func() {
		signalchan := make(chan os.Signal, 1)
		defer close(signalchan)
		signal.Notify(signalchan, syscall.SIGINT, syscall.SIGTERM)

		terminating := false
		for {
			signal := <-signalchan
			log.Warn(signal.String() + " recieved, stopping gracefully...")
			terminating = true
			messageFetcherCancel()

			if terminating {
				log.Error("recieved multiple exit signals, exiting immediately")
				os.Exit(130)
			}
		}
	}()

	darwinDatabaseConnection, err := darwindb.NewConnection(log.With(slog.String("source", "darwin.db")), databaseContext, c.Darwin.Database.URL)
	if err != nil {
		return fmt.Errorf("error connecting to the Darwin database: %w", err)
	}

	darwinKafkaConnection := darwinconnection.NewConnection(log.With(slog.String("source", "darwin.kafka")), kafkaContext, kafka.ReaderConfig{
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

	var fetcherGroup sync.WaitGroup
	go fetchMessages(log.With(slog.String("source", "darwin.fetcher")), &fetcherGroup, messageFetcherContext, darwinKafkaMessages, darwinKafkaConnection)

	var processorGroup sync.WaitGroup
	go processMessages(log.With(slog.String("source", "darwin.processor")), &processorGroup, darwinKafkaMessages, darwinDatabaseConnection)

	fetcherGroup.Wait()
	log.Info("waiting for processors to finish")
	processorGroup.Wait()

	log.Info("closing connections")
	if err := darwinKafkaConnection.Close(); err != nil {
		log.Error("error closing darwin kafka connection", slog.Any("error", err))
	}
	if err := darwinDatabaseConnection.Close(5 * time.Second); err != nil {
		log.Error("error closing darwin database connection", slog.Any("error", err))
	}
	return nil
}

// fetchMessages will run until the context is cancelled.
func fetchMessages(log *slog.Logger, wg *sync.WaitGroup, ctx context.Context, messages chan<- kafka.Message, connection kafkaConnection) {
	wg.Add(1)
	defer wg.Done()
	defer close(messages)
	for {
		message, err := connection.FetchMessage(ctx)
		if err != nil {
			if err == context.Canceled {
				log.Info("message fetching context cancelled, stopping fetcher")
				break
			}
			log.Warn("error fetching message", slog.Any("error", err))
			continue
		}
		messages <- message
	}
}

// processMessages will run until there are no more messages to process (the channel is closed and there are 0 messages remaining in it).
func processMessages(log *slog.Logger, wg *sync.WaitGroup, messages <-chan kafka.Message, db postgresqlConnection) {
	wg.Add(1)
	defer wg.Done()

	for message := range messages {

	}
}
