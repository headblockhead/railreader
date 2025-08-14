package processor

import (
	"errors"
	"log/slog"

	"github.com/headblockhead/railreader/darwin/decoder"
)

func (p *Processor) processServiceLoading(log *slog.Logger, msg *decoder.PushPortMessage, resp *decoder.Response, serviceLoading *decoder.ServiceLoading) error {
	return errors.New("Unimplemented")
}
