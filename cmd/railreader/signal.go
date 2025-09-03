package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func cancelOnSignal(cancel context.CancelFunc, log *slog.Logger) {
	signalchan := make(chan os.Signal, 1)
	defer close(signalchan)
	signal.Notify(signalchan, syscall.SIGINT, syscall.SIGTERM)

	alreadyTerminating := false
	for {
		signal := <-signalchan // block until a signal is received
		if alreadyTerminating {
			log.Warn("received multiple exit signals, exiting immediately")
			os.Exit(130)
		}
		alreadyTerminating = true
		log.Warn(signal.String() + " received, stopping gracefully...")
		cancel()
	}
}
