package railreader

import (
	"context"

	"github.com/segmentio/kafka-go"
)

type Ingester interface {
	FetchMessage(ctx context.Context) (kafka.Message, error)
	ProcessAndCommitMessage(msg kafka.Message) error
	Close() error
}
