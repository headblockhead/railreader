package interpreter

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/headblockhead/railreader/darwin/unmarshaller"
)

func (u UnitOfWork) InterpretPushPortMessage(pport unmarshaller.PushPortMessage) error {
	u.log.Debug("parsing new PushPortMessage timestamp")
	timestamp, err := time.Parse(time.RFC3339Nano, pport.Timestamp)
	if err != nil {
		return fmt.Errorf("failed to parse timestamp %q: %w", pport.Timestamp, err)
	}
	u.log.Debug("interpreting a PushPortMessage", slog.Time("timestamp", timestamp), slog.String("version", pport.Version))

	if pport.NewTimetableFiles != nil {
		// TODO: implement
		return errors.New("PushPortMessage contains NewTimetableFiles, which is not yet implemented")
	}
	if pport.StatusUpdate != nil {
		// TODO: implement
		return errors.New("PushPortMessage contains a StatusUpdate, which is not yet implemented")
	}
	if pport.UpdateResponse != nil {
		if err := u.interpretResponse(timestamp, false, pport.UpdateResponse); err != nil {
			return fmt.Errorf("failed to process UpdateResponse: %w", err)
		}
		return nil
	}
	if pport.SnapshotResponse != nil {
		if err := u.interpretResponse(timestamp, true, pport.SnapshotResponse); err != nil {
			return fmt.Errorf("failed to process SnapshotResponse: %w", err)
		}
		return nil
	}
	return errors.New("PushPortMessage was empty")
}

func (u UnitOfWork) interpretResponse(lastUpdated time.Time, snapshot bool, resp *unmarshaller.Response) error {
	u.log.Debug("interpreting a Response", slog.String("updateOrigin", resp.Source), slog.String("requestSourceSystem", resp.SourceSystem), slog.Bool("snapshot", snapshot))
	for _, schedule := range resp.Schedules {
		if err := interpretSchedule(u.log.With("RID", schedule.RID), u.messageID, lastUpdated, resp.Source, resp.SourceSystem, u.scheduleRepository, schedule); err != nil {
			return fmt.Errorf("failed to process Schedule %s: %w", schedule.RID, err)
		}
	}
	return nil
}
