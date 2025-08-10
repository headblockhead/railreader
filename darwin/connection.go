package darwin

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"

	"github.com/headblockhead/railreader/darwin/db"
	"github.com/headblockhead/railreader/darwin/decoder"
)

type Connection struct {
	connectionContext  context.Context
	fetcherContext     context.Context
	reader             *kafka.Reader
	databaseConnection *db.Connection
}

func NewConnection(connectionContext context.Context, fetcherContext context.Context, dbConnection *db.Connection, bootstrapServer string, groupID string, username string, password string) *Connection {
	return &Connection{
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
		databaseConnection: dbConnection,
	}
}

func (dc *Connection) Close() error {
	return dc.reader.Close()
}

// FetchKafkaMessage blocks until a message is available, or the fetcherContext is cancelled.
func (dc *Connection) FetchKafkaMessage(log *slog.Logger) (msg kafka.Message, err error) {
	if err := dc.fetcherContext.Err(); err != nil {
		return msg, fmt.Errorf("context error: %w", err)
	}
	log.Debug("blocking until Kafka message fetched")
	msg, err = dc.reader.FetchMessage(dc.fetcherContext)
	if err != nil {
		return msg, fmt.Errorf("failed to fetch a Kafka message: %w", err)
	}
	log.Debug("fetched a Kafka message")
	return msg, nil
}

func (dc *Connection) ProcessKafkaMessage(log *slog.Logger, msg kafka.Message) error {
	capsule, err := newMessageCapsule(log, msg)
	if err != nil {
		return fmt.Errorf("failed to create message capsule: %w", err)
	}
	log = log.With(slog.String("messageID", capsule.MessageID))

	pport, err := decoder.NewPushPortMessage(bytes.NewReader([]byte(capsule.Bytes)))
	if err != nil {
		return fmt.Errorf("failed to decode message bytes: %w", err)
	}
	log.Debug("unmarshalled into a PushPort message")

	// TODO: move somewhere else?

	if err := dc.commitKafkaMessage(log, msg); err != nil {
		return fmt.Errorf("failed to commit Kafka message: %w", err)
	}
	log.Debug("processed Kafka message")
	return nil
}

// messageCapsule is the raw JSON structure as received from the Rail Data Marketplace's Kafka topic.
// It contains a ridiculous amount of completely useless data and is practically fully undocumented, so I ignore everything but the message data inside, and the message's ID.
type messageCapsule struct {
	MessageID string `json:"messageID"`
	Bytes     string `json:"bytes"`
}

func newMessageCapsule(log *slog.Logger, msg kafka.Message) (*messageCapsule, error) {
	var c messageCapsule
	if err := json.Unmarshal(msg.Value, &c); err != nil {
		return nil, fmt.Errorf("failed to unmarshal kafka message: %w", err)
	}
	log.Debug("unmarshaled message capsule", slog.String("messageID", c.MessageID))
	return &c, nil
}

func (dc *Connection) commitKafkaMessage(log *slog.Logger, msg kafka.Message) error {
	if err := dc.reader.CommitMessages(dc.connectionContext, msg); err != nil {
		return fmt.Errorf("failed to commit Kafka message: %w", err)
	}
	log.Debug("committed to Kafka")
	return nil
}
