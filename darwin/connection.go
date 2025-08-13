package darwin

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"

	"github.com/headblockhead/railreader/darwin/db"
	"github.com/headblockhead/railreader/darwin/processor"
)

type Connection struct {
	log                *slog.Logger
	connectionContext  context.Context
	fetcherContext     context.Context
	reader             *kafka.Reader
	databaseConnection *db.Connection
	processor          *processor.Processor
}

func NewConnection(log *slog.Logger, connectionContext context.Context, fetcherContext context.Context, dbConnection *db.Connection, bootstrapServer string, groupID string, username string, password string) *Connection {
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
		databaseConnection: dbConnection,
	}
}

func (dc *Connection) Close() error {
	dc.log.Info("closing connection...")
	return dc.reader.Close()
}

// FetchKafkaMessage blocks until a message is available, or the fetcherContext is cancelled.
func (dc *Connection) FetchKafkaMessage() (*kafka.Message, error) {
	if err := dc.fetcherContext.Err(); err != nil {
		return nil, fmt.Errorf("context error: %w", err)
	}
	dc.log.Debug("blocking until Kafka message fetched")
	msg, err := dc.reader.FetchMessage(dc.fetcherContext)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch a Kafka message: %w", err)
	}
	dc.log.Debug("fetched Kafka message")
	return &msg, nil
}

func (dc *Connection) ProcessAndCommitKafkaMessage(msg *kafka.Message) error {
	if err := dc.processor.processKafkaMessage(msg); err != nil {
		return fmt.Errorf("failed to process Kafka message: %w", err)
	}
	dc.log.Debug("processed Kafka message")
	if err := dc.reader.CommitMessages(dc.connectionContext, *msg); err != nil {
		return fmt.Errorf("failed to commit Kafka message: %w", err)
	}
	dc.log.Debug("committed Kafka message")
	return nil
}
