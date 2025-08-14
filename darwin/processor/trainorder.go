package processor

import (
	"errors"
	"log/slog"

	"github.com/headblockhead/railreader/darwin/decoder"
)

func (p *Processor) processTrainOrder(log *slog.Logger, msg *decoder.PushPortMessage, resp *decoder.Response, trainOrder *decoder.TrainOrder) error {
	return errors.New("Unimplemented")
}
