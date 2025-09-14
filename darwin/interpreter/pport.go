package interpreter

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/headblockhead/railreader/darwin/repositories"
	"github.com/headblockhead/railreader/darwin/unmarshaller"
)

// when updating this, don't forget to update the version referenced in the unmarshaller package and its tests.
var expectedPushPortVersion = "18.0"

func (u UnitOfWork) InterpretPushPortMessage(pport unmarshaller.PushPortMessage) error {
	u.log.Debug("interpreting a PushPortMessage")
	if pport.Version != expectedPushPortVersion {
		u.log.Warn("PushPortMessage version does not match expected version", slog.String("expected_version", expectedPushPortVersion), slog.String("actual_version", pport.Version))
	}
	location, err := time.LoadLocation("Europe/London")
	if err != nil {
		return fmt.Errorf("failed to load time location: %w", err)
	}
	timestamp, err := time.Parse(time.RFC3339Nano, pport.Timestamp)
	if err != nil {
		return fmt.Errorf("failed to parse timestamp %q: %w", pport.Timestamp, err)
	}
	if err := u.pportMessageRepository.Insert(repositories.PPortMessage{
		MessageID:      u.messageID,
		SentAt:         timestamp.In(location),
		LastReceivedAt: time.Now().In(location),
		Version:        pport.Version,
	}); err != nil {
		return fmt.Errorf("failed to insert message record: %w", err)
	}

	if pport.NewTimetableFiles != nil {
		if err := u.handleNewTimetableFiles(pport.NewTimetableFiles); err != nil {
			return fmt.Errorf("failed to handle NewTimetableFiles: %w", err)
		}
		return nil
	}
	if pport.StatusUpdate != nil {
		// TODO: implement
		return errors.New("PushPortMessage contains a StatusUpdate, which is not yet implemented")
	}
	if pport.UpdateResponse != nil {
		if err := u.interpretResponse(false, pport.UpdateResponse); err != nil {
			return fmt.Errorf("failed to process UpdateResponse: %w", err)
		}
		return nil
	}
	if pport.SnapshotResponse != nil {
		if err := u.interpretResponse(true, pport.SnapshotResponse); err != nil {
			return fmt.Errorf("failed to process SnapshotResponse: %w", err)
		}
		return nil
	}
	return errors.New("PushPortMessage was empty")
}

var referenceTimetableVersionExension = "_v8.xml.gz"
var referenceRefVersionExtension = "_ref_v4.xml.gz"

func (u UnitOfWork) handleNewTimetableFiles(tf *unmarshaller.TimetableFiles) error {
	u.log.Debug("handling NewTimetableFiles")
	// todo: check each message for specific version extension, and limit to only those versions we support
	//timtableFile, err := u.ref.Get(tf.TimeTableId + referenceTimetableVersionExension)

	return nil
}

func (u UnitOfWork) interpretResponse(snapshot bool, resp *unmarshaller.Response) error {
	u.log.Debug("interpreting a Response", slog.Bool("snapshot", snapshot))
	var interpretedResponse repositories.Response
	interpretedResponse.MessageID = u.messageID
	interpretedResponse.Snapshot = snapshot
	if resp.Source != nil && *resp.Source != "" {
		u.log.Debug("source is set")
		interpretedResponse.Source = resp.Source
	}
	if resp.SourceSystem != nil && *resp.SourceSystem != "" {
		u.log.Debug("source system is set")
		interpretedResponse.SourceSystem = resp.SourceSystem
	}
	if resp.RequestID != nil && *resp.RequestID != "" {
		u.log.Debug("request ID is set")
		interpretedResponse.RequestID = resp.RequestID
	}
	if err := u.responseRepository.Insert(interpretedResponse); err != nil {
		return fmt.Errorf("failed to insert response record: %w", err)
	}
	for _, schedule := range resp.Schedules {
		if err := interpretSchedule(u.log.With("RID", schedule.RID), u.messageID, u.scheduleRepository, schedule); err != nil {
			return fmt.Errorf("failed to process Schedule %s: %w", schedule.RID, err)
		}
	}
	return nil
}
