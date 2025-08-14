package processor

import (
	"errors"
	"log/slog"

	"github.com/headblockhead/railreader/darwin/decoder"
)

func (p *Processor) processHeadcodeChange(log *slog.Logger, msg *decoder.PushPortMessage, resp *decoder.Response, change *decoder.HeadcodeChange) error {
	return errors.New("Unimplemented")
}
