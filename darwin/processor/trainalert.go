package processor

import (
	"errors"
	"log/slog"

	"github.com/headblockhead/railreader/darwin/decoder"
)

func (p *Processor) processTrainAlert(log *slog.Logger, msg *decoder.PushPortMessage, resp *decoder.Response, trainAlert *decoder.TrainAlert) error {
	return errors.New("Unimplemented")
}
