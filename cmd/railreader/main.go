package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/alecthomas/kong"
	"github.com/headblockhead/railreader/darwin"
	darwindb "github.com/headblockhead/railreader/darwin/db"
	"github.com/segmentio/kafka-go"
)

type CLI struct {
	Serve ServeCommand `cmd:"serve" aliases:"run" default:"withargs"`
}

type ServeCommand struct {
	Darwin struct {
		Server      string `group:"Darwin Push Port connection:" env:"DARWIN_SERVER" required:"" help:"Kafka server hostname and port."`
		GroupID     string `group:"Darwin Push Port connection:" env:"DARWIN_GROUPID" required:"" help:"Kafka consumer group ID." name:"group"`
		Username    string `group:"Darwin Push Port connection:" env:"DARWIN_USERNAME" required:"" help:"Kafka username."`
		Password    string `group:"Darwin Push Port connection:" env:"DARWIN_PASSWORD" required:"" help:"Kafka password."`
		DatabaseURL string `group:"Darwin Push Port connection:" env:"DARWIN_DATABASE_URL" required:"" help:"PostgreSQL database URL."`
	} `embed:"" prefix:"darwin-"`

	JSONOutput bool   `default:"false" short:"j" help:"Output logs as JSON instead of plaintext."`
	LogLevel   string `default:"info" enum:"debug,info,warn,error" help:"Minimum severity of logs required for them to be output."`
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
		log := getLogger("error", true)
		log.Error("error", slog.Any("error", err))
		os.Exit(1)
	}
}

var serverTerminating bool = false

type connectionWithKafka interface {
	FetchKafkaMessage() (*kafka.Message, error)
	ProcessAndCommitKafkaMessage(msg *kafka.Message) error
	Close() error
}

func (c ServeCommand) Run() error {
	log := getLogger(c.LogLevel, c.JSONOutput)

	fetcherContext, fetcherCancel := context.WithCancel(context.Background())

	signalchan := make(chan os.Signal, 1)
	signal.Notify(signalchan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for {
			<-signalchan
			if !serverTerminating {
				log.Warn("SIGINT/SIGTERM recieved, stopping gracefully...")
				serverTerminating = true
				fetcherCancel()
			} else {
				log.Error("recieved multiple SIGINT/SIGTERM signals, exiting immediately")
				os.Exit(130)
			}
		}
	}()

	darwinLog := log.With(slog.String("source", "darwin"))
	darwinDBLog := log.With(slog.String("source", "darwin.db"))

	var darwinConnection connectionWithKafka

	darwinDBLog.Debug("connecting")
	darwinDBContext, darwinDBCancel := context.WithCancel(context.Background())
	darwinDBConnection, err := darwindb.NewConnection(darwinDBContext, darwinDBLog, c.Darwin.DatabaseURL)
	if err != nil {
		darwinDBCancel()
		darwinDBLog.Error("error connecting", slog.Any("error", err))
		return err
	}
	darwinDBLog.Debug("connected")

	connectionContext, connectionCancel := context.WithCancel(context.Background())
	darwinConnection = darwin.NewConnection(darwinLog, connectionContext, fetcherContext, darwinDBConnection, c.Darwin.Server, c.Darwin.GroupID, c.Darwin.Username, c.Darwin.Password)

	darwinKafkaMessages := make(chan *kafka.Message, 32)
	var darwinProcessorGroup sync.WaitGroup

	//for range runtime.NumCPU() {
	for range 1 {
		darwinProcessorGroup.Add(1)
		go func() {
			defer darwinProcessorGroup.Done()
			processKafkaMessages(darwinLog, darwinConnection, darwinKafkaMessages)
		}()
	}

	// FetchKafkaMessages will run forever until the fetcherContext is canceled.
	fetchKafkaMessages(darwinLog, darwinConnection, darwinKafkaMessages)

	// Close the channel to indicate no more messages will be added by the fetchers.
	close(darwinKafkaMessages)

	log.Debug("stopped fetching new messages, waiting for processors to finish")
	darwinProcessorGroup.Wait()
	log.Debug("all processors finished, closing connections")

	connectionCancel()
	if err := darwinConnection.Close(); err != nil {
		darwinLog.Error("error closing connection", slog.Any("error", err))
	}

	darwinDBCancel()
	if err := darwinDBConnection.Close(5 * time.Second); err != nil {
		darwinDBLog.Error("error closing connection", slog.Any("error", err))
	}

	return nil
}

func fetchKafkaMessages(log *slog.Logger, c connectionWithKafka, messageChannel chan *kafka.Message) {
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
