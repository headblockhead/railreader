package db

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
)

type Connection struct {
	log        *slog.Logger
	url        string
	context    context.Context
	connection *pgx.Conn
}

func NewConnection(context context.Context, log *slog.Logger, url string) (*Connection, error) {
	conn, err := pgx.Connect(context, url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	return &Connection{
		log:        log,
		url:        url,
		context:    context,
		connection: conn,
	}, nil
}

func (c *Connection) Close(timeout time.Duration) error {
	c.log.Info("closing connection...")
	connectionCloseContext, connectionCloseCancel := context.WithTimeout(c.context, timeout)
	defer connectionCloseCancel()
	return c.connection.Close(connectionCloseContext)
}
