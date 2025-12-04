package main

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func connectToDatabase(ctx context.Context, log *slog.Logger, url string) (*pgxpool.Pool, error) {
	srcDriver, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return nil, fmt.Errorf("failed to create iofs for migrations: %w", err)
	}
	m, err := migrate.NewWithSourceInstance("iofs", srcDriver, url)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrate instance: %w", err)
	}
	log.Debug("connected migration tool to database")

	err = m.Up()
	if err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Debug("database schema is already up to date")
		} else {
			return nil, fmt.Errorf("failed to apply migrations: %w", err)
		}
	} else {
		log.Debug("migrated to the latest schema")
	}

	srcerr, dberr := m.Close()
	if srcerr != nil {
		return nil, fmt.Errorf("failed to close migrate connection due to an error closing the source: %w", srcerr)
	}
	if dberr != nil {
		return nil, fmt.Errorf("failed to close migrate connection due to an error closing the database: %w", dberr)
	}
	log.Debug("disconnected migration tool from database")

	dbpool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	log.Debug("connected pgx to database")

	return dbpool, nil
}
