package processor

import (
	"errors"
	"log/slog"

	"github.com/headblockhead/railreader/darwin/decoder"
)

func (p *Processor) processFormationLoading(log *slog.Logger, msg *decoder.PushPortMessage, resp *decoder.Response, formationLoading *decoder.FormationLoading) error {
	return errors.New("Unimplemented")
}
