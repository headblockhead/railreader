package interpreter

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/headblockhead/railreader/darwin/filegetter"
	"github.com/headblockhead/railreader/darwin/repository"
	"github.com/headblockhead/railreader/database"
	"github.com/jackc/pgx/v5"
)

type UnitOfWork struct {
	ctx       context.Context
	log       *slog.Logger
	messageID string
	tx        pgx.Tx
	fg        filegetter.FileGetter

	referenceRepository    repository.Reference
	pportMessageRepository repository.PPortMessage
	responseRepository     repository.Response
	scheduleRepository     repository.Schedule
}

func NewUnitOfWork(ctx context.Context, log *slog.Logger, messageID string, db database.Database, fg filegetter.FileGetter) (unit UnitOfWork, err error) {
	tx, err := db.BeginTx()
	if err != nil {
		err = fmt.Errorf("failed to begin new transaction: %w", err)
		return
	}
	referenceRepository := repository.NewPGXReference(ctx, log.With(slog.String("repository", "Reference")), tx)
	pportMessageRepository := repository.NewPGXPPortMessage(ctx, log.With(slog.String("repository", "PPortMessage")), tx)
	responseRepository := repository.NewPGXResponse(ctx, log.With(slog.String("repository", "Response")), tx)
	scheduleRepository := repository.NewPGXSchedule(ctx, log.With(slog.String("repository", "Schedule")), tx)
	unit = UnitOfWork{
		ctx:       ctx,
		log:       log,
		messageID: messageID,
		tx:        tx,
		fg:        fg,

		referenceRepository:    referenceRepository,
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
