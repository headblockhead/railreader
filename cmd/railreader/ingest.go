package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"runtime"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
)

type IngestCommand struct {
	DatabaseURL          string `env:"POSTGRESQL_URL" required:"" help:"PostgreSQL database URL to store data in."`
	SFTPWorkingDirectory string `env:"SFTP_WORKING_DIRECTORY" help:"Directory that the railreader SFTP server writes its files to." type:"existingdir" required:""`
	Darwin               struct {
		Kafka struct {
			Brokers           []string      `env:"DARWIN_KAFKA_BROKERS" default:"pkc-z4p1v0.europe-west2.gcp.confluent.cloud:9092" help:"Kafka broker(s) to connect to."`
			Topic             string        `env:"DARWIN_KAFKA_TOPIC" default:"prod-1010-Darwin-Train-Information-Push-Port-IIII2_0-XML" help:"Kafka topic for Darwin's XML feed."`
			Group             string        `env:"DARWIN_KAFKA_GROUP" required:""`
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

type messageFetcherCommitter interface {
	Close() error
	FetchMessage(ctx context.Context) (kafka.Message, error)
	CommitMessage(msg kafka.Message) error
}

type messageHandler interface {
	Handle(msg kafka.Message) error
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

	darwinFetcherCommiter, darwinMessageHandler, err := c.newDarwin(c.log.With(slog.String("source", "darwin")), dbpool)
	if err != nil {
		return fmt.Errorf("error setting up darwin connection: %w", err)
	}
	darwinKafkaMessages := make(chan kafka.Message, c.Darwin.QueueSize)

	messageFetcherContext, messageFetcherCancel := context.WithCancel(context.Background())
	go onSignal(c.log, func() {
		messageFetcherCancel()
	})
	var fetcherGroup sync.WaitGroup
	fetcherGroup.Go(func() {
		fetchMessages(messageFetcherContext, c.log.With(slog.String("source", "darwin"), slog.String("process", "fetcher")), darwinKafkaMessages, darwinFetcherCommiter)
		close(darwinKafkaMessages)
	})
	var handlerGroup sync.WaitGroup
	threadCount := runtime.NumCPU()
	for i := range threadCount {
		handlerGroup.Go(func() {
			handleMessages(c.log.With(slog.String("source", "darwin"), slog.String("process", "handler"), slog.Int("goroutine", i)), darwinKafkaMessages, darwinFetcherCommiter, darwinMessageHandler)
		})
	}

	// The fetcher group will run until messageFetcherContext is cancelled (when the program receives an exit signal).
	fetcherGroup.Wait()
	c.log.Info("waiting to finish processing all queued messages")
	handlerGroup.Wait()

	c.log.Info("closing connections")
	if err := darwinFetcherCommiter.Close(); err != nil {
		c.log.Error("error closing darwin kafka connection", slog.Any("error", err))
	}

	return nil
}

// fetchMessages will run until the context is cancelled.
func fetchMessages(ctx context.Context, log *slog.Logger, messages chan<- kafka.Message, fetcherCommitter messageFetcherCommitter) {
	log.Debug("starting message fetcher loop")
	for {
		message, err := fetcherCommitter.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				break
			}
			log.Error("error fetching message", slog.Any("error", err))
			continue
		}
		messages <- message
	}
	log.Debug("stopped fetching new messages")
}

// handleMessages will run until there are no more messages to handle (the channel is closed and there are 0 messages remaining in it).
func handleMessages(log *slog.Logger, messages <-chan kafka.Message, fetcherCommitter messageFetcherCommitter, handler messageHandler) {
	log.Debug("starting message handler loop")
	for msg := range messages {
		if err := handler.Handle(msg); err != nil {
			log.Error("error handling message", slog.Any("error", err))
			continue
		}
		if err := fetcherCommitter.CommitMessage(msg); err != nil {
			log.Error("error committing message", slog.Any("error", err))
			continue
		}
	}
	log.Debug("all queued messages handled")
}
