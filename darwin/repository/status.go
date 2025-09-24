package repository

import (
	"context"
	"log/slog"
	"time"

	"github.com/headblockhead/railreader/database"
	"github.com/jackc/pgx/v5"
)

type StatusRow struct {
	MessageID   string    `db:"message_id"`
	Code        string    `db:"code"`
	ReceivedAt  time.Time `db:"received_at"`
	Description string    `db:"description"`
}
type Status interface {
	Insert(status StatusRow) error
}
type PGXStatus struct {
	ctx context.Context
	log *slog.Logger
	tx  pgx.Tx
}

func NewPGXStatus(ctx context.Context, log *slog.Logger, tx pgx.Tx) PGXStatus {
	return PGXStatus{ctx, log, tx}
}
func (mr PGXStatus) Insert(status StatusRow) error {
	mr.log.Debug("inserting StatusRow", slog.String("message_id", status.MessageID))
	return database.InsertIntoTable(mr.ctx, mr.tx, "message_status", status)
}
