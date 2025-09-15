package repository

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
)

type MessageXML interface {
	Insert(messageXML MessageXMLRow) error
}

type PGXMessageXML struct {
	ctx context.Context
	log *slog.Logger
	tx  pgx.Tx
}

func NewPGXMessageXML(ctx context.Context, log *slog.Logger, tx pgx.Tx) PGXMessageXML {
	log.Debug("creating new PGXMessageXML")
	return PGXMessageXML{
		ctx: ctx,
		log: log,
		tx:  tx,
	}
}

type MessageXMLRow struct {
	MessageID string
	Offset    int64
	XML       string
}

func (mr PGXMessageXML) Insert(messageXML MessageXMLRow) error {
	mr.log.Debug("inserting MessageXMLRow")
	if _, err := mr.tx.Exec(mr.ctx, `
		INSERT INTO message_xml
			VALUES (
				@message_id
				,@offset
				,@xml
			) 
			ON CONFLICT (message_id) DO
			NOTHING;
	`, pgx.StrictNamedArgs{
		"message_id": messageXML.MessageID,
		"offset":     messageXML.Offset,
		"xml":        messageXML.XML,
	}); err != nil {
		return fmt.Errorf("failed to insert: %w", err)
	}
	return nil
}
