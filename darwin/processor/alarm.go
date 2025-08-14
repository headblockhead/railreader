package processor

import (
	"errors"
	"log/slog"

	"github.com/headblockhead/railreader/darwin/decoder"
)

func (p *Processor) processAlarm(log *slog.Logger, msg *decoder.PushPortMessage, resp *decoder.Response, alarm *decoder.Alarm) error {
	return errors.New("Unimplemented")
}
