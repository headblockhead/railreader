package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/alecthomas/kong"
	"github.com/headblockhead/railreader/darwin"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
)

type CLI struct {
	Serve ServeCommand `group:"Server:" cmd:"serve" help:"Run the railreader server."`

	Interpret InterpretCommand `group:"Client:" cmd:"interpret" help:"Interpret data from a single message, and write it to the database."`
}

type InterpretCommand struct {
	Darwin InterpretDarwinCommand `cmd:"darwin" help:"Interpret a message from Darwin."`
}

type InterpretDarwinCommand struct {
	MessageID string `arg:"" help:"message_id of a message to (re-)interpret. This will be fetched from the database."`
	File      string `arg:"" help:"Path to a file containing a Darwin message to interpret. This takes precedence over providing a message_id."`
}

type ServeCommand struct {
	Darwin struct {
		Kafka struct {
			Host     string `group:"Darwin Push Port client:" env:"DARWIN_KAFKA_HOST" required:"" help:"Kafka server hostname and port."`
			GroupID  string `group:"Darwin Push Port client:" env:"DARWIN_KAFKA_GROUP" required:"" name:"group"`
			Topic    string `group:"Darwin Push Port client:" env:"DARWIN_KAFKA_TOPIC" required:""`
			Username string `group:"Darwin Push Port client:" env:"DARWIN_KAFKA_USERNAME" required:""`
			Password string `group:"Darwin Push Port client:" env:"DARWIN_KAFKA_PASSWORD" required:""`
		} `embed:"" prefix:"kafka."`

		QueueSize   int    `group:"Darwin Push Port client:" env:"DARWIN_MESSAGE_QUEUE_SIZE" default:"32" help:"Maximum number of incoming messages to queue for processing at once. This does not affect data integrity, but will affect memory usage, bandwidth usage on startup, and how long it will take for the server to cleanly exit."`
		DatabaseURL string `group:"Darwin Push Port client:" env:"DARWIN_POSTGRESQL_URL" required:"" help:"PostgreSQL database URL to store Darwin data in."`
	} `embed:"" prefix:"darwin."`

	Socket struct {
		Enable   bool   `group:"Socket:" env:"SOCKET_ENABLE" default:"true" help:"Enable local control of the database via an unauthenticated socket. Access can be limited using file permissions."`
		Location string `group:"Socket:" env:"SOCKET_LOCATION" help:"Path of the socket file to create."`
		Mode     string `group:"Socket:" env:"SOCKET_MODE" default:"600" help:"File mode for the socket. The socket file is owned by the user running the program."`
	} `embed:"" prefix:"socket."`

	/* HTTP struct {*/
	/*Host string `group:"HTTP Server:" env:"HTTP_HOST" default:":8080" help:"HTTP server hostname and port."`*/
	/*} `embed:"" prefix:"http."`*/

	Logging struct {
		Level string `enum:"debug,info,warn,error" default:"info"`
		Type  string `enum:"json,console" default:"console"`
	} `embed:"" prefix:"logging."`
}

func getLogger(logLevel string, JSONOutput bool) *slog.Logger {
	var level slog.Level
	switch logLevel {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}
	var log *slog.Logger
	if JSONOutput {
		log = slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
			Level: level,
		}))
	} else {
		log = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: level,
		}))
	}
	return log
}

func main() {
	var cli CLI
	kctx := kong.Parse(&cli, kong.Description("Middleman between National Rail and Network Rail's datafeeds, and your project!"), kong.UsageOnError())

	if err := kctx.Run(); err != nil {
		// Assume logs should be output in JSON if the option cannot be obtained from the CLI.
		log := getLogger("error", true)
		log.Error("error", slog.Any("error", err))
		os.Exit(1)
	}
}

func (c ServeCommand) Run() error {
	log := getLogger(c.Logging.Level, c.Logging.Type == "json")

	messageFetcherContext, messageFetcherCancel := context.WithCancel(context.Background())

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

	var darwinConnection kafkaConnection

	darwinDBContext, darwinDBCancel := context.WithCancel(context.Background())
	darwinDBConnection, err := darwindb.NewConnection(darwinDBContext, darwinDBLog, c.Darwin.DatabaseURL)
	if err != nil {
		darwinDBCancel()
		darwinDBLog.Error("error connecting", slog.Any("error", err))
		return err
	}

	darwinProcessor := darwinprocessor.NewProcessor(darwinProcessorLog, darwinDBConnection)

	darwinKafkaReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{c.Darwin.Server},
		GroupID: c.Darwin.GroupID,
		Topic:   c.Darwin.Topic,
		Dialer: &kafka.Dialer{
			Timeout:   10 * time.Second,
			DualStack: true,
			SASLMechanism: plain.Mechanism{
				Username: c.Darwin.Username,
				Password: c.Darwin.Password,
			},
			TLS: &tls.Config{},
		},
	})

	connectionContext, connectionCancel := context.WithCancel(context.Background())
	darwinConnection = darwin.NewConnection(darwinConnectionLog, connectionContext, messageFetcherContext, darwinProcessor, darwinKafkaReader)

	darwinKafkaMessages := make(chan *kafka.Message, 32)
	var darwinProcessorGroup sync.WaitGroup

	// for range runtime.NumCPU() {
	for range 1 {
		darwinProcessorGroup.Add(1)
		go func() {
			defer darwinProcessorGroup.Done()
			processKafkaMessages(darwinConnectionLog, darwinConnection, darwinKafkaMessages)
		}()
	}

	// FetchKafkaMessages will run forever until the fetcherContext is canceled.
	fetchKafkaMessages(darwinConnectionLog, darwinConnection, darwinKafkaMessages)

	// Close the channel to indicate no more messages will be added by the fetchers.
	close(darwinKafkaMessages)

	log.Debug("stopped fetching new messages, waiting for processors to finish")
	darwinProcessorGroup.Wait()
	log.Debug("all processors finished, closing connections")

	connectionCancel()
	if err := darwinConnection.Close(); err != nil {
		darwinConnectionLog.Error("error closing connection", slog.Any("error", err))
	}

	darwinDBCancel()
	if err := darwinDBConnection.Close(5 * time.Second); err != nil {
		darwinDBLog.Error("error closing connection", slog.Any("error", err))
	}

	return nil
}

func fetchKafkaMessagesIntoChannel(log *slog.Logger, messageChannel chan *kafka.Message) {
	for {
		msg, err := c.FetchKafkaMessage()
		if err != nil {
			if errors.Is(err, context.Canceled) {
				log.Debug("context canceled while fetching Kafka message")
				break
			}
			log.Error("error fetching Kafka message", slog.Any("error", err))
			continue
		}
		messageChannel <- msg
	}
}

func processKafkaMessages(log *slog.Logger, c connectionWithKafka, messageChannel chan *kafka.Message) {
	for msg := range messageChannel {
		if serverTerminating {
			log.Debug("program terminating, processing remaining Kafka messages", slog.Int("remaining", len(messageChannel)))
		}
		err := c.ProcessAndCommitKafkaMessage(msg)
		if err != nil {
			log.Error("error processing Kafka message", slog.Any("error", err))
			continue
		}
	}
}

type Connection struct {
	log *slog.Logger

	connectionContext context.Context
	fetcherContext    context.Context
	reader            *kafka.Reader
}

func NewConnection(log *slog.Logger, connectionContext context.Context, fetcherContext context.Context, reader *kafka.Reader) *Connection {
	return &Connection{
		log:               log,
		connectionContext: connectionContext,
		fetcherContext:    fetcherContext,
		reader:            reader,
	}
}

func (dc *Connection) Close() error {
	dc.log.Info("closing connection...")
	return dc.reader.Close()
}

// FetchMessage blocks until a message is available, or the fetcherContext is cancelled.
func (dc *Connection) FetchMessage() (*kafka.Message, error) {
}

func (dc *Connection) CommitMessage(msg *kafka.Message) error {
	dc.log.Debug("commiting Kafka message", slog.Int64("offset", msg.Offset))
	if err := dc.reader.CommitMessages(dc.connectionContext, *msg); err != nil {
		return fmt.Errorf("failed to commit Kafka message: %w", err)
	}
	dc.log.Debug("committed Kafka message", slog.Int64("offset", msg.Offset))
	return nil
}
