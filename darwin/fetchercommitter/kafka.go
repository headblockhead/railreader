package fetchercommitter

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/segmentio/kafka-go"
)

type Kafka struct {
	ctx    context.Context
	cancel context.CancelCauseFunc
	log    *slog.Logger
	reader *kafka.Reader
}

func NewKafka(ctx context.Context, log *slog.Logger, readerConfig kafka.ReaderConfig) Kafka {
	ctx, cancel := context.WithCancelCause(ctx)
	return Kafka{
		ctx:    ctx,
		cancel: cancel,
		log:    log,
		reader: kafka.NewReader(readerConfig),
	}
}

func (c Kafka) Close() error {
	c.log.Debug("closing connection")
	defer c.cancel(errors.New("connection closed"))
	if err := c.reader.Close(); err != nil {
		return fmt.Errorf("failed to close Kafka reader: %w", err)
	}
	c.log.Debug("closed connection")
	return nil
}

// FetchMessage blocks until a message is available, or the provided context is cancelled.
func (c Kafka) FetchMessage(ctx context.Context) (msg kafka.Message, err error) {
	if err = ctx.Err(); err != nil {
		return
	}
	c.log.Debug("fetching message")
	msg, err = c.reader.FetchMessage(ctx)
	if err != nil {
		err = fmt.Errorf("failed to fetch a message: %w", err)
		return
	}
	c.log.Info("fetched message", slog.Int64("offset", msg.Offset))
	return
}

func (c Kafka) CommitMessage(msg kafka.Message) error {
	if err := c.ctx.Err(); err != nil {
		return fmt.Errorf("context error: %w", err)
	}
	c.log.Debug("commiting message", slog.Int64("offset", msg.Offset))
	if err := c.reader.CommitMessages(c.ctx, msg); err != nil {
		return fmt.Errorf("failed to commit message: %w", err)
	}
	c.log.Info("committed message", slog.Int64("offset", msg.Offset))
	return nil
}
