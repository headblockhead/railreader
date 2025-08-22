package interpreter

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/headblockhead/railreader/darwin/database"
	"github.com/jackc/pgx/v5"
)

type UnitOfWork struct {
	ctx       context.Context
	log       *slog.Logger
	messageID string
	tx        pgx.Tx

	scheduleRepository scheduleRepository
}

func NewUnitOfWork(ctx context.Context, log *slog.Logger, messageID string, db database.Database) (unit UnitOfWork, err error) {
	tx, err := db.BeginTx()
	if err != nil {
		err = fmt.Errorf("failed to begin new transaction: %w", err)
		return
	}
	scheduleRepository := database.NewPGXScheduleRepository(ctx, log, tx)
	unit = UnitOfWork{
		ctx:       ctx,
		log:       log,
		messageID: messageID,
		tx:        tx,

		scheduleRepository: scheduleRepository,
	}
	return
}

func (u *UnitOfWork) Commit() error {
	if err := u.tx.Commit(u.ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

func (u *UnitOfWork) Rollback() error {
	if err := u.tx.Rollback(u.ctx); err != nil {
		return fmt.Errorf("failed to rollback transaction: %w", err)
	}
	return nil
}

type scheduleRepository interface {
	Insert(schedule *database.Schedule) error
}
