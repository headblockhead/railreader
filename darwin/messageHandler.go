package darwin

import (
	"context"
	"encoding/json"
	"io/fs"
	"log/slog"

	"github.com/headblockhead/railreader/darwin/interpreter"
	"github.com/headblockhead/railreader/darwin/unmarshaller"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/segmentio/kafka-go"
)

type MessageHandler struct {
	log    *slog.Logger
	ctx    context.Context
	dbpool *pgxpool.Pool
	fs     fs.FS
}

func NewMessageHandler(ctx context.Context, log *slog.Logger, dbpool *pgxpool.Pool, fs fs.FS) MessageHandler {
	return MessageHandler{
		ctx:    ctx,
		log:    log,
		dbpool: dbpool,
		fs:     fs,
	}
}

// messageCapsule is the raw JSON structure as received from the Rail Data Marketplace's Kafka topic.
// The JSON itself has a lot of useless data, so I cherry-pick out what I want.
type messageCapsule struct {
	MessageID  string `json:"messageID"`
	Properties struct {
		PushPortSequence struct {
			String string `json:"string"`
		} `json:"PushPortSequence"`
	} `json:"properties"`
	XML string `json:"bytes"`
}

func (m MessageHandler) Handle(msg kafka.Message) error {
	var capsule messageCapsule
	if err := json.Unmarshal(msg.Value, &capsule); err != nil {
		return err
	}
	log := m.log.With(slog.String("messageID", capsule.MessageID))
	if err := insertMessageCapsule(m.ctx, log, m.dbpool, msg.Offset, capsule); err != nil {
		return err
	}
	pport, err := unmarshaller.NewPushPortMessage(capsule.XML)
	if err != nil {
		return err
	}
	if err := interpretPushPortMessage(m.ctx, log, m.dbpool, m.fs, capsule.MessageID, pport); err != nil {
		return err
	}
	return nil
}

func insertMessageCapsule(ctx context.Context, log *slog.Logger, dbpool *pgxpool.Pool, offset int64, capsule messageCapsule) error {
	u, err := interpreter.NewUnitOfWork(ctx, log, dbpool, nil, &capsule.MessageID, nil)
	if err != nil {
		return err
	}
	if err := u.InsertMessageXMLRecord(interpreter.MessageXMLRecord{
		ID:            capsule.MessageID,
		KafkaOffset:   offset,
		PPortSequence: capsule.Properties.PushPortSequence.String,
		XML:           capsule.XML,
	}); err != nil {
		_ = u.Rollback()
		return err
	}
	if err := u.Commit(); err != nil {
		_ = u.Rollback()
		return err
	}
	return nil
}

func interpretPushPortMessage(ctx context.Context, log *slog.Logger, dbpool *pgxpool.Pool, fs fs.FS, messageID string, pport unmarshaller.PushPortMessage) error {
	u, err := interpreter.NewUnitOfWork(ctx, log, dbpool, fs, &messageID, nil)
	if err != nil {
		return err
	}
	if err = u.InterpretPushPortMessage(pport); err != nil {
		_ = u.Rollback()
		return err
	}
	if err := u.Commit(); err != nil {
		_ = u.Rollback()
		return err
	}
	return nil
}
