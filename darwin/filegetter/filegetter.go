package filegetter

import (
	"io"
)

type FileGetter interface {
	Get(filepath string) (io.ReadCloser, error)
	GetLatestPathWithSuffix(suffix string) (string, error)
}
