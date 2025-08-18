package darwin

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/segmentio/kafka-go"
)

type Connection struct {
	log *slog.Logger

	connectionContext context.Context
	fetcherContext    context.Context
	reader            *kafka.Reader
}

func NewConnection(log *slog.Logger, connectionContext context.Context, fetcherContext context.Context, reader *kafka.Reader) *Connection {
	return &Connection{
		log:               log,
		connectionContext: connectionContext,
		fetcherContext:    fetcherContext,
		reader:            reader,
	}
}

func (dc *Connection) Close() error {
	dc.log.Info("closing connection...")
	return dc.reader.Close()
}

// FetchMessage blocks until a message is available, or the fetcherContext is cancelled.
func (dc *Connection) FetchMessage() (*kafka.Message, error) {
	if err := dc.fetcherContext.Err(); err != nil {
		return nil, fmt.Errorf("context error: %w", err)
	}
	dc.log.Debug("blocking until Kafka message fetched")
	msg, err := dc.reader.FetchMessage(dc.fetcherContext)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch a Kafka message: %w", err)
	}
	dc.log.Debug("fetched Kafka message", slog.Int64("offset", msg.Offset))
	return &msg, nil
}

func (dc *Connection) CommitMessage(msg *kafka.Message) error {
	dc.log.Debug("commiting Kafka message", slog.Int64("offset", msg.Offset))
	if err := dc.reader.CommitMessages(dc.connectionContext, *msg); err != nil {
		return fmt.Errorf("failed to commit Kafka message: %w", err)
	}
	dc.log.Debug("committed Kafka message", slog.Int64("offset", msg.Offset))
	return nil
}
