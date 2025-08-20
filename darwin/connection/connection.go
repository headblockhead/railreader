package connection

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/segmentio/kafka-go"
)

type Connection struct {
	log     *slog.Logger
	context context.Context
	cancel  context.CancelCauseFunc
	reader  *kafka.Reader
}

func New(log *slog.Logger, ctx context.Context, readerConfig kafka.ReaderConfig) *Connection {
	ctx, cancel := context.WithCancelCause(ctx)
	return &Connection{
		log:     log,
		context: ctx,
		cancel:  cancel,
		reader:  kafka.NewReader(readerConfig),
	}
}

func (c *Connection) Close() error {
	c.log.Info("closing connection")
	defer c.cancel(errors.New("connection closed"))
	if err := c.reader.Close(); err != nil {
		return fmt.Errorf("failed to close Kafka reader: %w", err)
	}
	c.log.Debug("connection closed successfully")
	return nil
}

// FetchMessage blocks until a message is available, or the provided context is cancelled.
func (c *Connection) FetchMessage(ctx context.Context) (kafka.Message, error) {
	if err := ctx.Err(); err != nil {
		return kafka.Message{}, fmt.Errorf("context error: %w", err)
	}
	c.log.Debug("blocking until message fetched")
	msg, err := c.reader.FetchMessage(ctx)
	if err != nil {
		return kafka.Message{}, fmt.Errorf("failed to fetch a message: %w", err)
	}
	c.log.Info("fetched message", slog.Int64("offset", msg.Offset))
	return msg, nil
}

func (c *Connection) CommitMessage(msg kafka.Message) error {
	if err := c.context.Err(); err != nil {
		return fmt.Errorf("context error: %w", err)
	}
	c.log.Debug("commiting message", slog.Int64("offset", msg.Offset))
	if err := c.reader.CommitMessages(c.context, msg); err != nil {
		return fmt.Errorf("failed to commit message: %w", err)
	}
	c.log.Debug("committed message", slog.Int64("offset", msg.Offset))
	return nil
}
