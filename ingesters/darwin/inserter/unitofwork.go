package inserter

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
	conn     *pgxpool.Conn
	tx       pgx.Tx
	batch    *pgx.Batch
	fs       fs.ReadDirFS

	messageID   *string
	timetableID *string
}

func NewUnitOfWork(ctx context.Context, log *slog.Logger, dbpool *pgxpool.Pool, fs fs.ReadDirFS, messageID *string, timetableID *string) (*UnitOfWork, error) {
	conn, err := dbpool.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire database connection: %w", err)
	}
	tx, err := conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to begin new transaction: %w", err)
	}
	log.Debug("created new transcation for unit of work")

	londonTime, err := time.LoadLocation("Europe/London")
	if err != nil {
		return nil, err
	}

	return &UnitOfWork{
		ctx:         ctx,
		log:         log,
		timezone:    londonTime,
		conn:        conn,
		tx:          tx,
		batch:       &pgx.Batch{},
		fs:          fs,
		messageID:   messageID,
		timetableID: timetableID,
	}, nil
}

func (u *UnitOfWork) Commit() error {
	batchResult := u.tx.SendBatch(u.ctx, u.batch)
	if err := batchResult.Close(); err != nil {
		return fmt.Errorf("failed to send batch: %w", err)
	}
	u.log.Debug("sent SQL batch to database")
	if err := u.tx.Commit(u.ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	u.log.Debug("committed unit of work transaction")
	u.conn.Release()
	return nil
}

func (u *UnitOfWork) Rollback() error {
	if err := u.tx.Rollback(u.ctx); err != nil {
		return fmt.Errorf("failed to rollback transaction: %w", err)
	}
	u.log.Debug("rolled back unit of work transaction")
	u.conn.Release()
	return nil
}
