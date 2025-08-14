package processor

import (
	"errors"
	"log/slog"

	"github.com/headblockhead/railreader/darwin/decoder"
)

func (p *Processor) processAssociation(log *slog.Logger, msg *decoder.PushPortMessage, resp *decoder.Response, association *decoder.Association) error {
	return errors.New("Unimplemented")
}
