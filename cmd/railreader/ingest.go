package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/headblockhead/railreader/darwin"
	darwinconn "github.com/headblockhead/railreader/darwin/connection"
	darwindb "github.com/headblockhead/railreader/darwin/database"
	"github.com/jackc/pgx/v5"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
)

type IngestCommand struct {
	Darwin struct {
		Kafka struct {
			Brokers           []string      `env:"DARWIN_KAFKA_BROKERS" default:"pkc-z3p1v0.europe-west2.gcp.confluent.cloud:9092" help:"Kafka broker(s) to connect to."`
			Topic             string        `env:"DARWIN_KAFKA_TOPIC" default:"prod-1010-Darwin-Train-Information-Push-Port-IIII2_0-XML" help:"Kafka topic to subscribe to for Darwin's XML feed."`
			Group             string        `env:"DARWIN_KAFKA_GROUP" required:"" help:"Consumer group."`
			UsernameFile      string        `env:"DARWIN_KAFKA_USERNAME" required:"" help:"File containing the consumer username in plaintext." type:"filecontent"`
			PasswordFile      string        `env:"DARWIN_KAFKA_PASSWORD_FILE" required:"" help:"File containing the consumer password in plaintext." type:"filecontent"`
			ConnectionTimeout time.Duration `env:"DARWIN_KAFKA_CONNECTION_TIMEOUT" default:"30s" help:"Timeout for connecting to the Kafka broker."`
		} `embed:"" prefix:"kafka."`
		Database struct {
			URL string `env:"DARWIN_POSTGRESQL_URL" required:"" help:"PostgreSQL database URL to store Darwin data in."`
		} `embed:"" prefix:"database."`
		QueueSize int `env:"DARWIN_QUEUE_SIZE" default:"32" help:"Maximum number of incoming messages to queue for processing at once. This does not affect data integrity, but will affect memory usage, bandwidth usage on startup, and how long it will take for the server to cleanly exit."`
	} `embed:"" prefix:"darwin."`

	Logging struct {
		Level  string `enum:"debug,info,warn,error" env:"LOG_LEVEL" default:"warn"`
		Format string `enum:"json,console" env:"LOG_FORMAT" default:"json"`
	} `embed:"" prefix:"log."`
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

func (c IngestCommand) Run() error {
	log := getLogger(c.Logging.Level, c.Logging.Format == "json")

	messageFetcherContext, messageFetcherCancel := context.WithCancel(context.Background())
	go onSignal(log, func() {
		messageFetcherCancel()
	})

	darwinDatabase, err := darwindb.New(context.Background(), log.With(slog.String("source", "darwin.database")), c.Darwin.Database.URL)
	if err != nil {
		return fmt.Errorf("error connecting to darwin database: %w", err)
	}

	kafkaContext := context.Background()
	darwinKafkaConnection := darwinconn.New(kafkaContext, log.With(slog.String("source", "darwin.connection")), kafka.ReaderConfig{
		Brokers: c.Darwin.Kafka.Brokers,
		GroupID: c.Darwin.Kafka.Group,
		Topic:   c.Darwin.Kafka.Topic,
		Dialer: &kafka.Dialer{
			Timeout:   c.Darwin.Kafka.ConnectionTimeout,
			DualStack: true,
			SASLMechanism: plain.Mechanism{
				Username: c.Darwin.Kafka.UsernameFile,
				Password: c.Darwin.Kafka.PasswordFile,
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
			log.Error("error fetching message", slog.Any("error", err))
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
