package processor

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/headblockhead/railreader/darwin/db"
	"github.com/headblockhead/railreader/darwin/decoder"
	"github.com/segmentio/kafka-go"
)

type Processor struct {
	log                *slog.Logger
	databaseConnection *db.Connection
}

func NewProcessor(log *slog.Logger, dbConnection *db.Connection) *Processor {
	return &Processor{
		log:                log,
		databaseConnection: dbConnection,
	}
}

// messageCapsule is the raw JSON structure as received from the Rail Data Marketplace's Kafka topic.
// It contains a ridiculous amount of completely useless data and is practically fully undocumented, so I ignore everything but the message data inside, and the message's ID.
type messageCapsule struct {
	MessageID string `json:"messageID"`
	XML       string `json:"bytes"`
}

func (p *Processor) ProcessKafkaMessage(msg *kafka.Message) error {
	if msg == nil {
		return errors.New("Kafka.Message is nil")
	}

	var capsule messageCapsule
	if err := json.Unmarshal(msg.Value, &capsule); err != nil {
		return fmt.Errorf("failed to unmarshal message JSON into messageCapsule: %w", err)
	}
	messageLog := p.log.With(slog.String("messageID", capsule.MessageID))
	messageLog.Debug("unmarshalled message JSON into a messageCapsule")

	pport, err := decoder.NewPushPortMessage(capsule.XML)
	if err != nil {
		return fmt.Errorf("failed to unmarshal messageCapsule XML into PushPortMessage: %w", err)
	}
	messageLog.Debug("unmarshalled messageCapsule XML into a PushPortMessage")

	if err := p.processPushPortMessage(messageLog, pport); err != nil {
		return fmt.Errorf("failed to process PushPortMessage: %w", err)
	}
	return nil
}
