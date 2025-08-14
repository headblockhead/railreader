package processor

import (
	"errors"
	"log/slog"

	"github.com/headblockhead/railreader/darwin/decoder"
)

func (p *Processor) processFormation(log *slog.Logger, msg *decoder.PushPortMessage, resp *decoder.Response, formation *decoder.FormationsOfService) error {
	return errors.New("Unimplemented")
}
