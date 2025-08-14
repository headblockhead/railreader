package processor

import (
	"errors"
	"log/slog"

	"github.com/headblockhead/railreader/darwin/decoder"
)

func (p *Processor) processDeactivation(log *slog.Logger, msg *decoder.PushPortMessage, resp *decoder.Response, deactivation *decoder.Deactivation) error {
	return errors.New("Unimplemented")
}
