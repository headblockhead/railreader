package darwin

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/headblockhead/railreader/darwin/database"
	"github.com/headblockhead/railreader/darwin/interpreter"
	"github.com/headblockhead/railreader/darwin/unmarshaller"
	"github.com/segmentio/kafka-go"
)

type MessageHandler struct {
	log *slog.Logger
	ctx context.Context
	db  database.Database
}

func NewMessageHandler(ctx context.Context, log *slog.Logger, db database.Database) MessageHandler {
	return MessageHandler{
		ctx: ctx,
		log: log,
		db:  db,
	}
}

// messageCapsule is the raw JSON structure as received from the Rail Data Marketplace's Kafka topic.
// It contains a ridiculous amount of completely useless data and is practically fully undocumented, so I ignore everything but the message data inside, and the message's ID.
type messageCapsule struct {
	MessageID string `json:"messageID"`
	XML       string `json:"bytes"`
}

// when updating this, don't forget to update the version referenced in the unmarshaller package and its tests.
var expectedPushPortVersion = "18.0"

func (m MessageHandler) Handle(msg kafka.Message) error {
	m.log.Debug("unmarshalling a new message's JSON into a messageCapsule (ID currently unknown)")
	var capsule messageCapsule
	if err := json.Unmarshal(msg.Value, &capsule); err != nil {
		return fmt.Errorf("failed to unmarshal the new message's JSON into messageCapsule: %w", err)
	}
	log := m.log.With(slog.String("message_id", capsule.MessageID))
	log.Debug("unmarshalled new message's JSON into a messageCapsule! (ID is now known)")
	if err := insertMessageCapsule(m.ctx, log, m.db, capsule); err != nil {
		return fmt.Errorf("failed to insert messageCapsule into database: %w", err)
	}
	log.Debug("creating a new PushPortMessage from the messageCapsule's XML")
	pport, err := unmarshaller.NewPushPortMessage(capsule.XML)
	if err != nil {
		return fmt.Errorf("failed to create new PushPortMessage: %w", err)
	}
	if pport.Version != expectedPushPortVersion {
		log.Warn("PushPortMessage version does not match expected version", slog.String("expected_version", expectedPushPortVersion), slog.String("actual_version", pport.Version))
	}
	if err := interpretPushPortMessage(m.ctx, log, m.db, capsule.MessageID, pport); err != nil {
		return fmt.Errorf("failed to interpret PushPortMessage: %w", err)
	}
	return nil
}

type MessageXMLRepository interface {
	Insert(message database.MessageXML) error
}

func insertMessageCapsule(ctx context.Context, log *slog.Logger, db database.Database, capsule messageCapsule) error {
	log.Debug("inserting a messageCapsule (as a database.MessageXML) into the database")
	messageXML := database.MessageXML{MessageID: capsule.MessageID, XML: capsule.XML}
	insertMessageXMLTx, err := db.BeginTx()
	if err != nil {
		_ = insertMessageXMLTx.Rollback(ctx)
		return fmt.Errorf("failed to begin MessageXML insertion transaction: %w", err)
	}
	mr := database.NewPGXMessageXMLRepository(ctx, log.With(slog.String("repository", "MessageXML")), insertMessageXMLTx)
	if err := mr.Insert(messageXML); err != nil {
		_ = insertMessageXMLTx.Rollback(ctx)
		return fmt.Errorf("failed to insert the MessageXML into the database: %w", err)
	}
	log.Debug("committing the MessageXML insertion transaction")
	if err := insertMessageXMLTx.Commit(ctx); err != nil {
		_ = insertMessageXMLTx.Rollback(ctx)
		return fmt.Errorf("failed to commit MessageXML insertion transaction: %w", err)
	}
	log.Debug("inserted a messageCapsule into the database")
	return nil
}

func interpretPushPortMessage(ctx context.Context, log *slog.Logger, db database.Database, messageID string, pport unmarshaller.PushPortMessage) error {
	log.Debug("interpreting a PushPortMessage")
	interpretPushPortMessageTx, err := db.BeginTx()
	if err != nil {
		return fmt.Errorf("failed to begin PushPortMessage interpretation transaction: %w", err)
	}
	log.Debug("creating a new UnitOfWork for interpreting the PushPortMessage")
	u := interpreter.NewUnitOfWork(
		log,
		messageID,
		database.NewPGXScheduleRepository(ctx, log.With(slog.String("repository", "Schedule")), interpretPushPortMessageTx),
	)
	if err = u.InterpretPushPortMessage(pport); err != nil {
		_ = interpretPushPortMessageTx.Rollback(ctx)
		return err
	}
	log.Debug("committing the PushPortMessage interpretation transaction")
	if err := interpretPushPortMessageTx.Commit(ctx); err != nil {
		_ = interpretPushPortMessageTx.Rollback(ctx)
		return fmt.Errorf("failed to commit PushPortMessage interpretation transaction: %w", err)
	}
	log.Debug("interpreted a PushPortMessage")
	return nil
}
