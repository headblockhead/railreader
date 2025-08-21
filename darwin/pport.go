package darwin

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/headblockhead/railreader/darwin/unmarshaller"
)

func (u *UnitOfWork) InterpretPushPortMessage(pport *unmarshaller.PushPortMessage) error {
	if pport == nil {
		return errors.New("PushPortMessage is nil")
	}

	timestamp, err := time.Parse(time.RFC3339Nano, pport.Timestamp)
	if err != nil {
		return fmt.Errorf("failed to parse timestamp %q: %w", pport.Timestamp, err)
	}
	u.log.Debug("processing PushPortMessage", slog.Time("timestamp", timestamp), slog.String("version", pport.Version))

	if pport.UpdateResponse != nil {
		if err := u.interpretResponse(timestamp, false, pport.UpdateResponse); err != nil {
			return fmt.Errorf("failed to process UpdateResponse: %w", err)
		}
		return nil
	}
	return errors.New("PushPortMessage does not contain any data")
}

func (u *UnitOfWork) interpretResponse(lastUpdated time.Time, snapshot bool, resp *unmarshaller.Response) error {
	u.log.Debug("processing Response", slog.String("updateOrigin", resp.Source), slog.String("requestSourceSystem", resp.SourceSystem), slog.Bool("snapshot", snapshot))
	for _, schedule := range resp.Schedules {
		if err := u.interpretSchedule(lastUpdated, resp.Source, resp.SourceSystem, &schedule); err != nil {
			return fmt.Errorf("failed to process Schedule %s: %w", schedule.RID, err)
		}
	}
	return nil
}
