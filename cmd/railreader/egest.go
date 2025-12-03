package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net"

	"github.com/headblockhead/railreader/egesters/grpc"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ServeCommand struct {
	DatabaseURL string `env:"POSTGRESQL_URL" required:"" help:"PostgreSQL database URL to store data in."`
	GRPCAddress string `env:"GRPC_ADDRESS" default:":50051"`

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

	grpcEgester, err := c.createGRPCServer(dbpool)
	if err != nil {
		return fmt.Errorf("error creating gRPC server: %w", err)
	}
	go onSignal(c.log, func() {
		grpcEgester.Close()
	})

	lis, err := net.Listen("tcp", c.GRPCAddress)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	return grpcEgester.Serve(lis)
}

func (c ServeCommand) createGRPCServer(dbpool *pgxpool.Pool) (*grpc.Egester, error) {
	return grpc.NewEgester(context.Background(), c.log.With(slog.String("source", "grpc")), dbpool)
}
