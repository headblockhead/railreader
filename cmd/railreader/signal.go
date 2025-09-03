package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func onSignal(log *slog.Logger, f func()) {
	signalchan := make(chan os.Signal, 1)
	defer close(signalchan)
	signal.Notify(signalchan, syscall.SIGINT, syscall.SIGTERM)

	alreadyTerminating := false
	for {
		signal := <-signalchan // block until a signal is received
		if alreadyTerminating {
			log.Error("received multiple exit signals, exiting immediately")
			os.Exit(130)
		}
		alreadyTerminating = true
		log.Info(signal.String() + " received, stopping gracefully...")
		f()
	}
}
