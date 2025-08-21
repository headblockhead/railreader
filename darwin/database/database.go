package database

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5"
)

type Database struct {
	ctx    context.Context
	cancel context.CancelCauseFunc
	log    *slog.Logger
	conn   *pgx.Conn
}

//go:embed migrations/*.sql
var migrationsFS embed.FS

// New connects to the postgres database, and automatically migrates the schema to the latest version.
func New(ctx context.Context, log *slog.Logger, url string) (*Database, error) {
	log.Debug("creating migrations iofs")
	srcDriver, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return nil, fmt.Errorf("failed to create iofs for migrations: %w", err)
	}
	log.Debug("connecting migration tool")
	m, err := migrate.NewWithSourceInstance("iofs", srcDriver, url)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrate instance: %w", err)
	}
	log.Debug("migrating to the latest schema")
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return nil, fmt.Errorf("failed to apply migrations: %w", err)
	}
	log.Debug("closing migration tool's connection")
	srcerr, dberr := m.Close()
	if srcerr != nil {
		return nil, fmt.Errorf("failed to close migrate connection due to an error closing the source: %w", srcerr)
	}
	if dberr != nil {
		return nil, fmt.Errorf("failed to close migrate connection due to an error closing the database: %w", dberr)
	}

	ctx, cancel := context.WithCancelCause(ctx)
	log.Debug("connecting pgx")
	conn, err := pgx.Connect(ctx, url)
	if err != nil {
		cancel(errors.New("failed to connect to database"))
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	log.Debug("connected pgx")

	return &Database{
		ctx:    ctx,
		cancel: cancel,
		log:    log,
		conn:   conn,
	}, nil
}

func (c *Database) Close(timeout time.Duration) error {
	c.log.Debug("closing connection", slog.String("timeout", timeout.String()))
	defer c.cancel(errors.New("connection closed"))
	ctx, cancel := context.WithTimeout(c.ctx, timeout)
	defer cancel()
	if err := c.conn.Close(ctx); err != nil {
		return fmt.Errorf("failed to close pgx connection: %w", err)
	}
	c.log.Debug("connection closed")
	return nil
}

func (c *Database) BeginTx() (pgx.Tx, error) {
	c.log.Debug("starting a new transaction")
	tx, err := c.conn.Begin(c.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	return tx, nil
}
