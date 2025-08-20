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
	"github.com/pkg/errors"
)

type Connection struct {
	log     *slog.Logger
	context context.Context
	cancel  context.CancelCauseFunc
	conn    *pgx.Conn
}

//go:embed migrations/*.sql
var migrationsFS embed.FS

// NewConnection connects to the postgres database, and automatically migrates the schema to the latest version.
func NewConnection(log *slog.Logger, ctx context.Context, url string) (*Connection, error) {
	log.Debug("creating iofs for migrations")
	srcDriver, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return nil, fmt.Errorf("failed to create iofs for migrations: %w", err)
	}
	log.Debug("connecting migrate")
	m, err := migrate.NewWithSourceInstance("iofs", srcDriver, url)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrate instance: %w", err)
	}
	log.Debug("migrating to the latest schema")
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return nil, fmt.Errorf("failed to apply migrations: %w", err)
	}
	log.Debug("closing migrate connection")
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
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	log.Debug("connected pgx")

	return &Connection{
		log:     log,
		context: ctx,
		cancel:  cancel,
		conn:    conn,
	}, nil
}

func (c *Connection) Close(timeout time.Duration) error {
	c.log.Debug("closing pgx connection", slog.String("timeout", timeout.String()))
	defer c.cancel(errors.New("connection closed"))
	// TODO: check if cancel is optional (below)
	ctx, cancel := context.WithTimeout(c.context, timeout)
	defer cancel()
	if err := c.conn.Close(ctx); err != nil {
		return fmt.Errorf("failed to close pgx connection: %w", err)
	}
	return nil
}

func (c *Connection) BeginTx() (pgx.Tx, error) {
	c.log.Debug("starting a new transaction")
	tx, err := c.conn.Begin(c.context)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	return tx, nil
}
