package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/alecthomas/kong"
)

type CLI struct {
	Ingest IngestCommand `cmd:"ingest" help:"Ingest data into the database from the message feeds."`
	SFTP   SFTPCommand   `cmd:"sftp" help:"Host an SFTP server that the Rail Data Marketplace can copy to."`
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

// onSignal blocks until a SIGINT or SIGTERM signal is received, then calls f() in a new goroutine.
// The context that is passed to onSignal should be cancelled when shutdown has completed successfully.
func onSignal(log *slog.Logger, ctx context.Context, f func()) {
	signalchan := make(chan os.Signal, 1)
	defer close(signalchan)
	signal.Notify(signalchan, syscall.SIGINT, syscall.SIGTERM)

	signal := <-signalchan // block until a signal is received
	fmt.Println()          // Always print a newline so the terminal prompt appears correctly.
	log.Info(signal.String() + " received, stopping gracefully...")
	go f()

	select {
	case <-signalchan:
		log.Error("received multiple exit signals, exiting immediately")
		os.Exit(1)
	case <-ctx.Done():
		log.Debug("graceful shutdown complete")
	}
}

func main() {
	var cli CLI
	kctx := kong.Parse(&cli, kong.Description("Middleman between various UK rail data sources and your project."), kong.UsageOnError())

	if err := kctx.Run(); err != nil {
		// Assume logs should be output in JSON if the option cannot be obtained from the CLI.
		log := getLogger("error", true)
		log.Error("error", slog.Any("error", err))
		os.Exit(1)
	}
}
