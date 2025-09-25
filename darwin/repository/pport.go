package repository

import (
	"context"
	"log/slog"
	"time"

	"github.com/headblockhead/railreader/database"
	"github.com/jackc/pgx/v5"
)

type PPortMessageRow struct {
	MessageID       string    `db:"message_id"`
	SentAt          time.Time `db:"sent_at"`
	FirstReceivedAt time.Time `db:"first_received_at"`
	Version         string    `db:"version"`
}
type PPortMessage interface {
	Insert(message PPortMessageRow) error
}
type PGXPPortMessage struct {
	ctx context.Context
	log *slog.Logger
	tx  pgx.Tx
}

func NewPGXPPortMessage(ctx context.Context, log *slog.Logger, tx pgx.Tx) PGXPPortMessage {
	return PGXPPortMessage{ctx, log, tx}
}

func (mr PGXPPortMessage) Insert(message PPortMessageRow) error {
	mr.log.Debug("inserting PPortMessageRow", slog.String("message_id", message.MessageID))
	return database.InsertIntoTable(mr.ctx, mr.tx, "messages", message)
}

type ResponseRow struct {
	MessageID    string  `db:"message_id"`
	IsSnapshot   bool    `db:"is_snapshot"`
	Source       *string `db:"source"`
	SourceSystem *string `db:"source_system"`
	RequestID    *string `db:"request_id"`
}
type Response interface {
	Insert(response ResponseRow) error
}
type PGXResponse struct {
	ctx context.Context
	log *slog.Logger
	tx  pgx.Tx
}

func NewPGXResponse(ctx context.Context, log *slog.Logger, tx pgx.Tx) PGXResponse {
	return PGXResponse{ctx, log, tx}
}

func (mr PGXResponse) Insert(repsonse ResponseRow) error {
	mr.log.Debug("inserting ResponseRow", slog.String("message_id", repsonse.MessageID))
	return database.InsertIntoTable(mr.ctx, mr.tx, "message_response", repsonse)
}
