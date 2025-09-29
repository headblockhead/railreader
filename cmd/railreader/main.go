package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/alecthomas/kong"
)

type CLI struct {
	Ingest IngestCommand `cmd:"ingest" help:"Ingest data into the database from the message feeds."`
	Serve  ServeCommand  `cmd:"serve" help:"Run the HTTP API server."`
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

// onSignal waits for a SIGINT or SIGTERM signal, then calls f().
func onSignal(log *slog.Logger, f func()) {
	signalchan := make(chan os.Signal, 1)
	defer close(signalchan)
	signal.Notify(signalchan, syscall.SIGINT, syscall.SIGTERM)

	alreadyTerminating := false
	for {
		signal := <-signalchan // block until a signal is received
		if alreadyTerminating {
			log.Error("received multiple exit signals, exiting immediately")
			os.Exit(1)
		}
		alreadyTerminating = true
		log.Info(signal.String() + " received, stopping gracefully...")
		f()
	}
}

func main() {
	var cli CLI
	kctx := kong.Parse(&cli, kong.Description("Middleman between National Rail's Darwin and your project!"), kong.UsageOnError())

	if err := kctx.Run(); err != nil {
		// Assume logs should be output in JSON if the option cannot be obtained from the CLI.
		log := getLogger("error", true)
		log.Error("error", slog.Any("error", err))
		os.Exit(1)
	}
}
