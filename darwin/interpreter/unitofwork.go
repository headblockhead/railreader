package interpreter

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/headblockhead/railreader/darwin/reference"
	"github.com/headblockhead/railreader/darwin/repositories"
	"github.com/headblockhead/railreader/database"
	"github.com/jackc/pgx/v5"
)

type UnitOfWork struct {
	ctx       context.Context
	log       *slog.Logger
	messageID string
	tx        pgx.Tx
	ref       reference.Connection

	pportMessageRepository repositories.PPortMessageRepository
	responseRepository     repositories.ResponseRepository
	scheduleRepository     repositories.ScheduleRepository
}

func NewUnitOfWork(ctx context.Context, log *slog.Logger, messageID string, db database.Database, ref reference.Connection) (unit UnitOfWork, err error) {
	tx, err := db.BeginTx()
	if err != nil {
		err = fmt.Errorf("failed to begin new transaction: %w", err)
		return
	}
	pportMessageRepository := repositories.NewPGXPPortMessageRepository(ctx, log.With(slog.String("repository", "PPortMessage")), tx)
	responseRepository := repositories.NewPGXResponseRepository(ctx, log.With(slog.String("repository", "Response")), tx)
	scheduleRepository := repositories.NewPGXScheduleRepository(ctx, log.With(slog.String("repository", "Schedule")), tx)
	unit = UnitOfWork{
		ctx:       ctx,
		log:       log,
		messageID: messageID,
		tx:        tx,
		ref:       ref,

		pportMessageRepository: pportMessageRepository,
		responseRepository:     responseRepository,
		scheduleRepository:     scheduleRepository,
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
