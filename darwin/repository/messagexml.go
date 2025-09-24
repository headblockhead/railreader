package repository

import (
	"context"
	"log/slog"

	"github.com/headblockhead/railreader/database"
	"github.com/jackc/pgx/v5"
)

type MessageXMLRow struct {
	MessageID   string `db:"message_id"`
	XML         string `db:"xml"`
	KafkaOffset int64  `db:"kafka_offset"`
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
	mr.log.Debug("inserting MessageXMLRow", slog.String("message_id", messageXML.MessageID), slog.Int64("offset", messageXML.KafkaOffset))
	return database.InsertIntoTable(mr.ctx, mr.tx, "message_xml", messageXML)
}
