package darwin

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/headblockhead/railreader/darwin/decoder"
	"github.com/segmentio/kafka-go"
)

func (dc *Connection) processKafkaMessage(msg *kafka.Message) error {
	dc.log.Debug("processing Kafka message")
	capsule, err := newMessageCapsule(dc.log, msg)
	if err != nil {
		return fmt.Errorf("failed to create message capsule: %w", err)
	}
	messageLog := dc.log.With(slog.String("messageID", capsule.MessageID))

	pport, err := decoder.NewPushPortMessage(bytes.NewReader([]byte(capsule.Bytes)))
	if err != nil {
		return fmt.Errorf("failed to decode message bytes: %w", err)
	}
	messageLog.Debug("unmarshalled into a PushPort message")

	if pport == nil {
		return errors.New("unmarshalled PushPort message is nil")
	}

	if err := dc.processPushPortMessage(messageLog, pport); err != nil {
		return fmt.Errorf("failed to process PushPortMessage: %w", err)
	}
	return nil
}

// messageCapsule is the raw JSON structure as received from the Rail Data Marketplace's Kafka topic.
// It contains a ridiculous amount of completely useless data and is practically fully undocumented, so I ignore everything but the message data inside, and the message's ID.
type messageCapsule struct {
	MessageID string `json:"messageID"`
	Bytes     string `json:"bytes"`
}

func newMessageCapsule(log *slog.Logger, msg *kafka.Message) (*messageCapsule, error) {
	var c messageCapsule
	if err := json.Unmarshal(msg.Value, &c); err != nil {
		return nil, fmt.Errorf("failed to unmarshal kafka message: %w", err)
	}
	log.Debug("unmarshalled message capsule", slog.String("messageID", c.MessageID))
	return &c, nil
}
