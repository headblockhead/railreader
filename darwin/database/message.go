package database

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
)

type MessageXMLRepository interface {
	Insert(messageXML MessageXML) error
}

type PGXMessageXMLRepository struct {
	ctx context.Context
	log *slog.Logger
	tx  pgx.Tx
}

func NewPGXMessageXMLRepository(ctx context.Context, log *slog.Logger, tx pgx.Tx) PGXMessageXMLRepository {
	log.Debug("creating new PGXMessageXMLRepository")
	return PGXMessageXMLRepository{
		ctx: ctx,
		log: log,
		tx:  tx,
	}
}

type MessageXML struct {
	MessageID string
	XML       string
}

func (mr PGXMessageXMLRepository) Insert(messageXML MessageXML) error {
	mr.log.Debug("inserting MessageXML")
	if _, err := mr.tx.Exec(mr.ctx, `
		INSERT INTO message_xml
			VALUES (
				@message_id,
				@xml
			) 
			ON CONFLICT (message_id) DO
			NOTHING;
	`, pgx.StrictNamedArgs{
		"message_id": messageXML.MessageID,
		"xml":        messageXML.XML,
	}); err != nil {
		return fmt.Errorf("failed to insert: %w", err)
	}
	return nil
}
