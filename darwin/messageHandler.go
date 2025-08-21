package darwin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/headblockhead/railreader/darwin/database"
	"github.com/headblockhead/railreader/darwin/unmarshaller"
	"github.com/segmentio/kafka-go"
)

type MessageHandler struct {
	log *slog.Logger
	ctx context.Context
	db  *database.Database
}

func NewMessageHandler(ctx context.Context, log *slog.Logger, db *database.Database) *MessageHandler {
	return &MessageHandler{
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

type MessageXMLRepository interface {
	Insert(message database.MessageXML) error
}

func (m *MessageHandler) Handle(msg *kafka.Message) error {
	m.log.Debug("handling a message")
	if msg == nil {
		return errors.New("kafka message is nil")
	}

	m.log.Debug("unmarshalling message JSON into a messageCapsule")
	var capsule messageCapsule
	if err := json.Unmarshal(msg.Value, &capsule); err != nil {
		return fmt.Errorf("failed to unmarshal message JSON into messageCapsule: %w", err)
	}

	log := m.log.With(slog.String("message_id", capsule.MessageID))

	log.Debug("beginning store transaction")
	storeMessageTx, err := m.db.BeginTx()
	if err != nil {
		return fmt.Errorf("failed to begin message store transaction: %w", err)
	}
	mr := database.NewPGXMessageXMLRepository(m.ctx, log, storeMessageTx)
	if err := mr.Insert(database.MessageXML{
		MessageID: capsule.MessageID,
		XML:       capsule.XML,
	}); err != nil {
		_ = storeMessageTx.Rollback(m.ctx)
		return fmt.Errorf("failed to store message in database: %w", err)
	}
	log.Debug("committing message store transaction")
	if err := storeMessageTx.Commit(m.ctx); err != nil {
		_ = storeMessageTx.Rollback(m.ctx)
		return fmt.Errorf("failed to commit message store transaction: %w", err)
	}

	log.Debug("unmarhalling message XML into a PushPortMessage")
	pport, err := unmarshaller.NewPushPortMessage(capsule.XML)
	if err != nil {
		return fmt.Errorf("failed to unmarshal message XML into PushPortMessage: %w", err)
	}

	log.Debug("beginning message execute transaction")
	executeMessageTx, err := m.db.BeginTx()
	if err != nil {
		return fmt.Errorf("failed to begin message execute transaction: %w", err)
	}

	u := UnitOfWork{
		log:                log,
		messageID:          capsule.MessageID,
		ScheduleRepository: database.NewPGXScheduleRepository(m.ctx, log, executeMessageTx),
	}

	if err := u.InterpretPushPortMessage(pport); err != nil {
		_ = executeMessageTx.Rollback(m.ctx)
		return fmt.Errorf("failed to process PushPortMessage: %w", err)
	}

	return nil
}
