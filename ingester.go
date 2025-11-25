package railreader

import (
	"context"
)

type Ingester[T any] interface {
	Fetch(context.Context) (T, error)
	ProcessAndCommit(T) error
	Close() error
}
