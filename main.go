package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"

	"github.com/alecthomas/kong"
	"github.com/segmentio/kafka-go"
)

type CLI struct {
	Serve ServeCommand `cmd:"serve" aliases:"run" default:"withargs"`
}

type ServeCommand struct {
	Darwin struct {
		Server   string `group:"Darwin Push Port connection:" env:"DARWIN_SERVER" required:"" help:"Kafka server hostname and port."`
		GroupID  string `group:"Darwin Push Port connection:" env:"DARWIN_GROUPID" required:"" help:"Kafka consumer group ID." name:"group"`
		Username string `group:"Darwin Push Port connection:" env:"DARWIN_USERNAME" required:"" help:"Kafka username."`
		Password string `group:"Darwin Push Port connection:" env:"DARWIN_PASSWORD" required:"" help:"Kafka password."`
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

var currentlyTerminating bool = false

func main() {
	var cli CLI
	kctx := kong.Parse(&cli, kong.Description("Middleman between National Rail and Network Rail's datafeeds, and your project!"), kong.UsageOnError())

	if err := kctx.Run(); err != nil {
		log := getLogger("error", true)
		log.Error("error", slog.Any("error", err))
		os.Exit(1)
	}
}

func (c ServeCommand) Run() error {
	log := getLogger(c.LogLevel, c.JSONOutput)
	connectionContext, connectionCancel := context.WithCancel(context.Background())
	fetcherContext, fetcherCancel := context.WithCancel(context.Background())

	signalchan := make(chan os.Signal, 1)
	signal.Notify(signalchan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for {
			<-signalchan
			if !currentlyTerminating {
				log.Warn("SIGINT/SIGTERM recieved, stopping gracefully...")
				currentlyTerminating = true
				fetcherCancel()
			} else {
				log.Error("recieved multiple SIGINT/SIGTERM signals, exiting immediately")
				os.Exit(130)
			}
		}
	}()

	dc := NewDarwinConnection(connectionContext, fetcherContext, log.With(slog.String("source", "darwin")), c.Darwin.Server, c.Darwin.GroupID, c.Darwin.Username, c.Darwin.Password)
	darwinKafkaMessages := make(chan kafka.Message, 64)

	var darwinProcessorGroup sync.WaitGroup

	for range runtime.NumCPU() {
		//for range 1 {
		darwinProcessorGroup.Add(1)
		go func() {
			defer darwinProcessorGroup.Done()
			processKafkaMessages(log, dc, darwinKafkaMessages)
		}()
	}

	FetchKafkaMessages(log, dc, darwinKafkaMessages)

	close(darwinKafkaMessages)
	darwinProcessorGroup.Wait()
	connectionCancel()
	if err := dc.Close(); err != nil {
		log.Error("error closing Darwin connection", slog.Any("error", err))
	}

	return nil
}

func FetchKafkaMessages(log *slog.Logger, dc *DarwinConnection, darwinKafkaMessages chan kafka.Message) {
	for {
		msg, err := dc.FetchKafkaMessage()
		if err != nil {
			if errors.Is(err, context.Canceled) {
				log.Debug("context canceled while fetching Kafka message")
				break
			}
			log.Error("error fetching Kafka message", slog.Any("error", err))
			continue
		}
		darwinKafkaMessages <- msg
	}
}

func processKafkaMessages(log *slog.Logger, dc *DarwinConnection, darwinKafkaMessages chan kafka.Message) {
	for msg := range darwinKafkaMessages {
		if currentlyTerminating {
			log.Debug("program terminating, processing remaining Kafka messages", slog.Int("remaining", len(darwinKafkaMessages)))
		}
		err := dc.ProcessKafkaMessage(msg)
		if err != nil {
			log.Error("error processing Kafka message", slog.Any("error", err))
			continue
		}
	}
}
