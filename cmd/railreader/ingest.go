package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/headblockhead/railreader"
	"github.com/headblockhead/railreader/ingesters/darwin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
)

type IngestCommand struct {
	DatabaseURL          string `env:"POSTGRESQL_URL" required:"" help:"PostgreSQL database URL to store data in."`
	SFTPWorkingDirectory string `env:"SFTP_WORKING_DIRECTORY" help:"Directory that the railreader SFTP server writes its files to." type:"existingdir" required:""`
	Darwin               struct {
		Kafka struct {
			Brokers           []string      `env:"DARWIN_KAFKA_BROKERS" help:"Kafka 'bootstrap server' to connect to." required:""`
			Topic             string        `env:"DARWIN_KAFKA_TOPIC" help:"Kafka topic for Darwin's XML feed." required:""`
			Group             string        `env:"DARWIN_KAFKA_GROUP" help:"Kafka 'consumer group'" required:""`
			Username          string        `env:"DARWIN_KAFKA_USERNAME" required:""`
			Password          string        `env:"DARWIN_KAFKA_PASSWORD" required:""`
			ConnectionTimeout time.Duration `env:"DARWIN_KAFKA_CONNECTION_TIMEOUT" default:"30s" help:"Timeout for connecting to the Kafka broker."`
		} `embed:"" prefix:"kafka."`
		QueueSize int `env:"DARWIN_QUEUE_SIZE" default:"32" help:"Maximum number of incoming messages to queue for processing at once."`
	} `embed:"" prefix:"darwin."`

	Logging struct {
		Level  string `enum:"debug,info,warn,error" env:"LOG_LEVEL" default:"warn"`
		Format string `enum:"json,console" env:"LOG_FORMAT" default:"console"`
	} `embed:"" prefix:"log."`

	log *slog.Logger `kong:"-"`
}

func (c IngestCommand) Run() error {
	c.log = getLogger(c.Logging.Level, c.Logging.Format == "json")

	var databaseContext, databaseCancel = context.WithCancel(context.Background())
	defer databaseCancel()
	dbpool, err := connectToDatabase(databaseContext, c.log.With(slog.String("process", "database")), c.DatabaseURL)
	if err != nil {
		return fmt.Errorf("error connecting to database: %w", err)
	}
	defer dbpool.Close()

	darwinIngester, err := c.createDarwinIngester(dbpool)
	if err != nil {
		return fmt.Errorf("error creating darwin ingester: %w", err)
	}
	darwinIngestQueue := make(chan kafka.Message, c.Darwin.QueueSize)

	messageFetcherContext, messageFetcherCancel := context.WithCancel(context.Background())
	go onSignal(c.log, messageFetcherCancel)

	var fetcherGroup sync.WaitGroup
	fetcherGroup.Go(func() {
		fetchMessages(messageFetcherContext, c.log.With(slog.String("source", "darwin"), slog.String("process", "fetcher")), darwinIngestQueue, darwinIngester)
		close(darwinIngestQueue)
	})
	var handlerGroup sync.WaitGroup
	cpus := runtime.NumCPU()
	for i := range cpus {
		handlerGroup.Go(func() {
			processAndCommitMessages(c.log.With(slog.String("source", "darwin"), slog.String("process", "handler"), slog.Int("goroutine", i)), darwinIngestQueue, darwinIngester)
		})
	}

	// The fetcher group will run until messageFetcherContext is cancelled (when the program receives an exit signal).
	fetcherGroup.Wait()
	c.log.Info("waiting to finish processing all queued messages")
	handlerGroup.Wait()

	c.log.Info("closing connections")
	err = darwinIngester.Close()
	if err != nil {
		c.log.Error("error closing darwin kafka connection", slog.Any("error", err))
	}

	return nil
}

func (c IngestCommand) createDarwinIngester(dbpool *pgxpool.Pool) (railreader.Ingester[kafka.Message], error) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: c.Darwin.Kafka.Brokers,
		GroupID: c.Darwin.Kafka.Group,
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

	root, err := os.OpenRoot(c.SFTPWorkingDirectory + "/darwin")
	if err != nil {
		return nil, err
	}

	var rootFS fs.FS = root.FS()
	var rdfs fs.ReadDirFS = rootFS.(fs.ReadDirFS)

	return darwin.NewIngester(context.Background(), c.log.With(slog.String("source", "darwin")), reader, dbpool, rdfs)
}

// fetchMessages will run until the context is cancelled.
func fetchMessages[T any](ctx context.Context, log *slog.Logger, messages chan<- T, ingester railreader.Ingester[T]) {
	for {
		msg, err := ingester.Fetch(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				break
			}
			log.Error("error fetching message", slog.Any("error", err))
			continue
		}
		messages <- msg
	}
}

// processAndCommitMessages will run until there are no more messages to handle (the channel is closed and there are 0 messages remaining in it).
func processAndCommitMessages[T any](log *slog.Logger, messages <-chan T, ingester railreader.Ingester[T]) {
	for msg := range messages {
		err := ingester.ProcessAndCommit(msg)
		if err != nil {
			log.Error("error processing and committing message", slog.Any("error", err))
		}
	}
}
