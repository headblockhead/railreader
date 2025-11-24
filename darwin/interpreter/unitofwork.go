package interpreter

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// A UnitOfWork represents a single transaction scope.
type UnitOfWork struct {
	ctx      context.Context
	log      *slog.Logger
	timezone *time.Location
	tx       pgx.Tx
	fs       fs.FS

	messageID   *string
	timetableID *string
}

func NewUnitOfWork(ctx context.Context, log *slog.Logger, dbpool *pgxpool.Pool, fs fs.FS, messageID *string, timetableID *string) (unit *UnitOfWork, err error) {
	log.Debug("creating new transaction for new unit of work")
	tx, err := dbpool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to begin new transaction: %w", err)
	}
	log.Debug("transaction created for new unit of work")

	londonTime, err := time.LoadLocation("Europe/London")
	if err != nil {
		return nil, err
	}

	return &UnitOfWork{
		ctx:         ctx,
		log:         log,
		timezone:    londonTime,
		tx:          tx,
		fs:          fs,
		messageID:   messageID,
		timetableID: timetableID,
	}, nil
}

func (u UnitOfWork) Commit() error {
	u.log.Debug("committing transaction for unit of work")
	if err := u.tx.Commit(u.ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	u.log.Debug("transaction committed for unit of work")
	return nil
}

func (u UnitOfWork) Rollback() error {
	u.log.Debug("rolling back transaction for unit of work")
	if err := u.tx.Rollback(u.ctx); err != nil {
		return fmt.Errorf("failed to rollback transaction: %w", err)
	}
	u.log.Debug("transaction rolled back for unit of work")
	return nil
}
