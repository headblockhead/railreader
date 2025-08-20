package main

import (
	"log/slog"
	"os"

	"github.com/alecthomas/kong"
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

	DryRun bool `help:"Do not write the message to the database."`

	Logging struct {
		Level string `enum:"debug,info,warn,error" default:"info"`
		Type  string `enum:"json,console" default:"console"`
	} `embed:"" prefix:"logging."`
	Socket struct {
		// TODO: grab location from SYSTEMD? /var/run/railreader/railreader.sock
		Location string `env:"SOCKET_LOCATION" help:"Path of the socket file to connect to."`
	} `embed:"" prefix:"socket."`
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
