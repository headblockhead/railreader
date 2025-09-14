package repositories

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
)

type PPortMessageRepository interface {
	Insert(message PPortMessage) error
}

type PGXPPortMessageRepository struct {
	ctx context.Context
	log *slog.Logger
	tx  pgx.Tx
}

func NewPGXPPortMessageRepository(ctx context.Context, log *slog.Logger, tx pgx.Tx) PGXPPortMessageRepository {
	log.Debug("creating new PGXMessageRepository")
	return PGXPPortMessageRepository{
		ctx: ctx,
		log: log,
		tx:  tx,
	}
}

type PPortMessage struct {
	MessageID      string
	SentAt         time.Time
	LastReceivedAt time.Time
	Version        string
}

func (mr PGXPPortMessageRepository) Insert(message PPortMessage) error {
	mr.log.Debug("inserting PPortMessage")
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

type ResponseRepository interface {
	Insert(response Response) error
}

type PGXResponseRepository struct {
	ctx context.Context
	log *slog.Logger
	tx  pgx.Tx
}

func NewPGXResponseRepository(ctx context.Context, log *slog.Logger, tx pgx.Tx) PGXResponseRepository {
	log.Debug("creating new PGXResponseRepository")
	return PGXResponseRepository{
		ctx: ctx,
		log: log,
		tx:  tx,
	}
}

type Response struct {
	MessageID    string
	Snapshot     bool
	Source       *string
	SourceSystem *string
	RequestID    *string
}

func (mr PGXResponseRepository) Insert(repsonse Response) error {
	mr.log.Debug("inserting Response")
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
