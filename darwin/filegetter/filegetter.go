package filegetter

import (
	"io"
)

type FileGetter interface {
	Get(filepath string) (io.ReadCloser, error)
	FindNewestWithSuffix(suffix string) (string, error)
}
