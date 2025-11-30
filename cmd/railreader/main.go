package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/alecthomas/kong"
	"github.com/phsym/console-slog"
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
		log = slog.New(console.NewHandler(os.Stderr, &console.HandlerOptions{
			Level: level,
		}))
	}
	return log
}

// onSignal blocks until a SIGINT or SIGTERM signal is received, then calls f() in a new goroutine.
func onSignal(log *slog.Logger, f func()) {
	signalchan := make(chan os.Signal, 1)
	defer close(signalchan)
	signal.Notify(signalchan, syscall.SIGINT, syscall.SIGTERM)

	// Block until a signal is received.
	signal := <-signalchan
	fmt.Print("\n") // an interactive terminal user may have pressed ^C (which is echoed) to produce a SIGINT, so add a newline for readability.
	log.Info(signal.String() + " received, stopping gracefully")
	go f()

	<-signalchan // block until a second signal is received.
	log.Error("received multiple exit signals, exiting immediately")
	os.Exit(1)
}

func main() {
	var cli CLI
	kctx := kong.Parse(&cli, kong.Description("Middleman between various UK rail data sources and your project."), kong.UsageOnError())

	if err := kctx.Run(); err != nil {
		log := getLogger("error", false)
		log.Error("error", slog.Any("error", err))
		os.Exit(1)
	}
}
