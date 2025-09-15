package repository

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
)

type PPortMessage interface {
	Insert(message PPortMessageRow) error
}

type PGXPPortMessage struct {
	ctx context.Context
	log *slog.Logger
	tx  pgx.Tx
}

func NewPGXPPortMessage(ctx context.Context, log *slog.Logger, tx pgx.Tx) PGXPPortMessage {
	log.Debug("creating new PGXMessage")
	return PGXPPortMessage{
		ctx: ctx,
		log: log,
		tx:  tx,
	}
}

type PPortMessageRow struct {
	MessageID      string
	SentAt         time.Time
	LastReceivedAt time.Time
	Version        string
}

func (mr PGXPPortMessage) Insert(message PPortMessageRow) error {
	mr.log.Debug("inserting PPortMessageRow")
	_, err := mr.tx.Exec(mr.ctx, `
		INSERT INTO messages
			VALUES (
				@message_id
				,@sent_at
				,@last_received_at
				,@version
			) 
			ON CONFLICT (message_id) DO
			UPDATE 
				SET 
					last_received_at = EXCLUDED.last_received_at;
	`, pgx.StrictNamedArgs{
		"message_id":       message.MessageID,
		"sent_at":          message.SentAt,
		"last_received_at": message.LastReceivedAt,
		"version":          message.Version,
	})
	return err
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
	log.Debug("creating new PGXResponse")
	return PGXResponse{
		ctx: ctx,
		log: log,
		tx:  tx,
	}
}

type ResponseRow struct {
	MessageID    string
	Snapshot     bool
	Source       *string
	SourceSystem *string
	RequestID    *string
}

func (mr PGXResponse) Insert(repsonse ResponseRow) error {
	mr.log.Debug("inserting ResponseRow")
	_, err := mr.tx.Exec(mr.ctx, `
		INSERT INTO message_response
			VALUES (
				@message_id
				,@snapshot
				,@source
				,@source_system
				,@request_id
			) 
			ON CONFLICT (message_id) DO
			NOTHING;
		`, pgx.StrictNamedArgs{
		"message_id":    repsonse.MessageID,
		"snapshot":      repsonse.Snapshot,
		"source":        repsonse.Source,
		"source_system": repsonse.SourceSystem,
		"request_id":    repsonse.RequestID,
	})
	return err
}
