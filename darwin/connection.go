package darwin

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
)

type Connection struct {
	log               *slog.Logger
	connectionContext context.Context
	fetcherContext    context.Context
	reader            *kafka.Reader
}

// MessageCapsule is the raw JSON structure as received from the Rail Data Marketplace's Kafka topic.
// It contains a ridiculous amount of completely useless data and is practically fully undocumented, so I ignore everything but the message data inside, and the message's ID.
type MessageCapsule struct {
	MessageID string `json:"messageID"`
	Bytes     string `json:"bytes"`
}

func NewConnection(connectionContext context.Context, fetcherContext context.Context, log *slog.Logger, bootstrapServer string, groupID string, username string, password string) *Connection {
	return &Connection{
		log:               log,
		connectionContext: connectionContext,
		fetcherContext:    fetcherContext,
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: []string{bootstrapServer},
			GroupID: groupID,
			Topic:   "prod-1010-Darwin-Train-Information-Push-Port-IIII2_0-XML",
			Dialer: &kafka.Dialer{
				Timeout:   10 * time.Second,
				DualStack: true,
				SASLMechanism: plain.Mechanism{
					Username: username,
					Password: password,
				},
				TLS: &tls.Config{},
			},
		}),
	}
}

func (dc *Connection) Close() error {
	dc.log.Info("closing connection...")
	return dc.reader.Close()
}

func (dc *Connection) FetchKafkaMessage() (msg kafka.Message, err error) {
	dc.log.Debug("waiting for a Kafka message to fetch...")
	if err := dc.fetcherContext.Err(); err != nil {
		return msg, fmt.Errorf("context error: %w", err)
	}
	msg, err = dc.reader.FetchMessage(dc.fetcherContext)
	if err != nil {
		return msg, fmt.Errorf("failed to fetch kafka message: %w", err)
	}
	var key struct {
		MessageID string `json:"messageID"`
	}
	if err := json.Unmarshal(msg.Key, &key); err != nil {
		return msg, fmt.Errorf("failed to unmarshal kafka message key: %w", err)
	}
	dc.log.Debug("received Kafka message", slog.String("messageID", key.MessageID))
	return msg, nil
}

func (dc *Connection) ProcessKafkaMessage(msg kafka.Message) error {
	if err := dc.connectionContext.Err(); err != nil {
		return fmt.Errorf("context error: %w", err)
	}

	var key struct {
		MessageID string `json:"messageID"`
	}
	if err := json.Unmarshal(msg.Key, &key); err != nil {
		return fmt.Errorf("failed to unmarshal kafka message key: %w", err)
	}

	log := dc.log.With(slog.String("messageID", string(key.MessageID)))

	log.Debug("unmarshaling Kafka message...")
	var c MessageCapsule
	if err := json.Unmarshal(msg.Value, &c); err != nil {
		return fmt.Errorf("failed to unmarshal kafka message: %w", err)
	}
	log.Debug("unmarshaled Kafka message")

	if err := dc.connectionContext.Err(); err != nil {
		return fmt.Errorf("context error: %w", err)
	}

	if err := dc.ProcessMessageCapsule(c); err != nil {
		return fmt.Errorf("failed to process message capsule: %w", err)
	}

	if err := dc.reader.CommitMessages(dc.connectionContext, msg); err != nil {
		return fmt.Errorf("failed to commit message: %w", err)
	}

	log.Debug("processed a message")

	return nil
}
