package db

import (
	"context"
	"embed"
	"fmt"
	"log/slog"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5"
)

type Connection struct {
	log           *slog.Logger
	url           string
	context       context.Context
	pgxConnection *pgx.Conn
}

//go:embed migrations/*.sql
var migrationsFS embed.FS

// NewConnection connects to the postgres database, and automatically migrates the schema to the latest version.
func NewConnection(context context.Context, log *slog.Logger, url string) (*Connection, error) {
	var conn Connection = Connection{
		log:     log,
		url:     url,
		context: context,
	}

	srcDriver, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return nil, fmt.Errorf("failed to create iofs for migrations: %w", err)
	}
	log.Debug("connecting migrate to the database")
	m, err := migrate.NewWithSourceInstance("iofs", srcDriver, url)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrate instance: %w", err)
	}
	log.Debug("connected migrate to the database")
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return nil, fmt.Errorf("failed to apply migrations: %w", err)
	}
	log.Debug("applied all upward migrations")
	srcerr, dberr := m.Close()
	if srcerr != nil {
		return nil, fmt.Errorf("failed to close migrate connection due to an error closing the source: %w", srcerr)
	}
	if dberr != nil {
		return nil, fmt.Errorf("failed to close migrate connection due to an error closing the database: %w", dberr)
	}
	log.Debug("closed migration instance")

	log.Debug("connecting pgx to the database")
	pgxConnection, err := pgx.Connect(context, url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	log.Debug("connected pgx to the database")
	conn.pgxConnection = pgxConnection

	return &conn, nil
}

func (c *Connection) Close(timeout time.Duration) error {
	c.log.Debug("closing pgx connection", slog.String("timeout", timeout.String()))
	connectionCloseContext, connectionCloseCancel := context.WithTimeout(c.context, timeout)
	defer connectionCloseCancel()
	if err := c.pgxConnection.Close(connectionCloseContext); err != nil {
		return fmt.Errorf("failed to close pgx connection: %w", err)
	}
	c.log.Debug("pgx connection closed")
	return nil
}
