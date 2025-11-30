package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/headblockhead/railreader/outputs/restful"
)

type ServeCommand struct {
	DatabaseURL          string `env:"POSTGRESQL_URL" required:"" help:"PostgreSQL database URL to store data in."`
	RESTfulListenAddress string `env:"RESTFUL_LISTEN_ADDRESS" default:"0.0.0.0:8034"`

	Logging struct {
		Level  string `enum:"debug,info,warn,error" env:"LOG_LEVEL" default:"warn"`
		Format string `enum:"json,console" env:"LOG_FORMAT" default:"console"`
	} `embed:"" prefix:"log."`

	log *slog.Logger `kong:"-"`
}

func (c ServeCommand) Run() error {
	c.log = getLogger(c.Logging.Level, c.Logging.Format == "json")

	var databaseContext, databaseCancel = context.WithCancel(context.Background())
	defer databaseCancel()
	dbpool, err := connectToDatabase(databaseContext, c.log.With(slog.String("process", "database")), c.DatabaseURL)
	if err != nil {
		return fmt.Errorf("error connecting to database: %w", err)
	}
	defer dbpool.Close()

	router := restful.NewRouter(c.log, dbpool)

	return router.Run()
}
