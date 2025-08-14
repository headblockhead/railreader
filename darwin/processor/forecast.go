package processor

import (
	"errors"
	"log/slog"

	"github.com/headblockhead/railreader/darwin/decoder"
)

func (p *Processor) processForecastTime(log *slog.Logger, msg *decoder.PushPortMessage, resp *decoder.Response, forecast *decoder.ForecastTime) error {
	return errors.New("Unimplemented")
}
