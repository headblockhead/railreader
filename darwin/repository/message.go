package repository

import (
	"context"
	"log/slog"

	"github.com/headblockhead/railreader/database"
	"github.com/jackc/pgx/v5"
)

type MessageXMLRow struct {
	MessageID string `db:"message_id"`
	Offset    int64  `db:"kafka_offset"`
	XML       string `db:"xml"`
}
type MessageXML interface {
	Insert(messageXML MessageXMLRow) error
}
type PGXMessageXML struct {
	ctx context.Context
	log *slog.Logger
	tx  pgx.Tx
}

func NewPGXMessageXML(ctx context.Context, log *slog.Logger, tx pgx.Tx) PGXMessageXML {
	return PGXMessageXML{ctx, log, tx}
}

func (mr PGXMessageXML) Insert(messageXML MessageXMLRow) error {
	mr.log.Debug("inserting MessageXMLRow", slog.String("message_id", messageXML.MessageID), slog.Int64("offset", messageXML.Offset))
	return database.InsertIntoTable(mr.ctx, mr.tx, "message_xml", messageXML)
}
