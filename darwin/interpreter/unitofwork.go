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

	messageRepository  database.MessageRepository
	responseRepository database.ResponseRepository
	scheduleRepository database.ScheduleRepository
}

func NewUnitOfWork(ctx context.Context, log *slog.Logger, messageID string, db database.Database) (unit UnitOfWork, err error) {
	tx, err := db.BeginTx()
	if err != nil {
		err = fmt.Errorf("failed to begin new transaction: %w", err)
		return
	}
	scheduleRepository := database.NewPGXScheduleRepository(ctx, log, tx)
	responseRepository := database.NewPGXResponseRepository(ctx, log, tx)
	messageRepository := database.NewPGXMessageRepository(ctx, log, tx)
	unit = UnitOfWork{
		ctx:       ctx,
		log:       log,
		messageID: messageID,
		tx:        tx,

		scheduleRepository: scheduleRepository,
		responseRepository: responseRepository,
		messageRepository:  messageRepository,
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
