package processor

import (
	"errors"
	"log/slog"

	"github.com/headblockhead/railreader/darwin/decoder"
)

func (p *Processor) processStationMessage(log *slog.Logger, msg *decoder.PushPortMessage, resp *decoder.Response, stationMessage *decoder.StationMessage) error {
	return errors.New("Unimplemented")
}
