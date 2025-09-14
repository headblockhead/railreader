package main

import (
	"log/slog"
	"os"

	"github.com/alecthomas/kong"
)

type CLI struct {
	Ingest IngestCommand `cmd:"ingest" help:"Ingest data into the database from the message feeds."`
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
	kctx := kong.Parse(&cli, kong.Description("Middleman between National Rail's datafeeds and your project!"), kong.UsageOnError())

	if err := kctx.Run(); err != nil {
		// Assume logs should be output in JSON if the option cannot be obtained from the CLI.
		log := getLogger("error", true)
		log.Error("error", slog.Any("error", err))
		os.Exit(1)
	}
}
