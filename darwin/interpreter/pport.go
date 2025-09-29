package interpreter

import (
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/headblockhead/railreader/darwin/filegetter"
	"github.com/headblockhead/railreader/darwin/repository"
	"github.com/headblockhead/railreader/darwin/unmarshaller"
)

// when updating this, don't forget to update the version referenced in the unmarshaller package and its tests.
var expectedPushPortVersion = "18.0"

// TODO: note to self: interpreters should probably return rows, or slices of rows, rather than running inserts themselves.
// still needs DB access to read stuff tho

func (u UnitOfWork) InterpretPushPortMessage(pport unmarshaller.PushPortMessage) error {
	u.log.Debug("interpreting a PushPortMessage")
	if pport.Version != expectedPushPortVersion {
		// Warn, but attempt to continue anyway.
		u.log.Warn("PushPortMessage version does not match expected version", slog.String("expected_version", expectedPushPortVersion), slog.String("actual_version", pport.Version))
	}
	location, err := time.LoadLocation("Europe/London")
	if err != nil {
		return fmt.Errorf("failed to load time location: %w", err)
	}
	timestamp, err := time.ParseInLocation(time.RFC3339Nano, pport.Timestamp, location)
	if err != nil {
		return fmt.Errorf("failed to parse timestamp %q: %w", pport.Timestamp, err)
	}
	if err := u.pportMessageRepository.Insert(repository.PPortMessageRow{
		MessageID:       u.messageID,
		SentAt:          timestamp,
		FirstReceivedAt: time.Now(),
		Version:         pport.Version,
	}); err != nil {
		return fmt.Errorf("failed to insert message record: %w", err)
	}

	if pport.NewFiles != nil {
		if err := u.handleNewFiles(pport.NewFiles); err != nil {
			return fmt.Errorf("failed to handle NewFiles: %w", err)
		}
		return nil
	}
	if pport.StatusUpdate != nil {
		if err := u.interpretStatus(pport.StatusUpdate); err != nil {
			return fmt.Errorf("failed to process StatusUpdate: %w", err)
		}
		return nil
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

func (u UnitOfWork) handleNewFiles(tf *unmarshaller.NewFiles) error {
	u.log.Debug("handling NewFiles")
	// Filter for specific version numbers of the files we care about.
	if strings.HasSuffix(tf.ReferenceFile, "_ref_v4.xml.gz") {
		GetUnmarshalAndInterpretFile(u.log, u.fg, tf.ReferenceFile, unmarshaller.NewReference, u.InterpretReference)
	}
	if strings.HasSuffix(tf.TimetableFile, "_v8.xml.gz") {
		GetUnmarshalAndInterpretFile(u.log, u.fg, tf.TimetableFile, unmarshaller.NewTimetable, u.InterpretTimetable)
	}
	return nil
}

func (u UnitOfWork) InterpretFromPath(path string) error {
	u.log.Debug("handling a filename", slog.String("path", path))
	if strings.HasSuffix(path, "_ref_v4.xml.gz") {
		return GetUnmarshalAndInterpretFile(u.log, u.fg, path, unmarshaller.NewReference, u.InterpretReference)
	}
	if strings.HasSuffix(path, "_v8.xml.gz") {
		return GetUnmarshalAndInterpretFile(u.log, u.fg, path, unmarshaller.NewTimetable, u.InterpretTimetable)
	}
	u.log.Info("filename does not match any known patterns, ignoring", slog.String("path", path))
	return nil
}

func GetUnmarshalAndInterpretFile[T any](log *slog.Logger, fg filegetter.FileGetter, path string, unmarshal func(string) (T, error), interpret func(T, string) error) error {
	log.Debug("getting file", slog.String("path", path))
	file, err := fg.Get(path)
	if err != nil {
		return fmt.Errorf("failed to get from filegetter: %w", err)
	}
	log.Debug("file gotten")
	reader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer reader.Close()
	contents, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("failed to read all of gzip reader: %w", err)
	}
	log.Debug("file read")
	data, err := unmarshal(string(contents))
	if err != nil {
		return err
	}
	log.Debug("file unmarshalled")
	if err := interpret(data, path); err != nil {
		return err
	}
	log.Debug("file interpreted")
	return nil
}

func (u UnitOfWork) interpretStatus(status *unmarshaller.Status) error {
	u.log.Debug("interpreting a Status")
	var row repository.StatusRow
	row.MessageID = u.messageID
	row.Code = string(status.Code)
	row.ReceivedAt = time.Now()
	row.Description = status.Description
	if err := u.statusRepository.Insert(row); err != nil {
		return err
	}
	return nil
}

func (u UnitOfWork) interpretResponse(snapshot bool, resp *unmarshaller.Response) error {
	u.log.Debug("interpreting a Response", slog.Bool("snapshot", snapshot))
	var row repository.ResponseRow
	row.MessageID = u.messageID
	row.IsSnapshot = snapshot
	if resp.Source != nil && *resp.Source != "" {
		u.log.Debug("source is set")
		row.Source = resp.Source
	}
	if resp.SourceSystem != nil && *resp.SourceSystem != "" {
		u.log.Debug("source system is set")
		row.SourceSystem = resp.SourceSystem
	}
	if resp.RequestID != nil && *resp.RequestID != "" {
		u.log.Debug("request ID is set")
		row.RequestID = resp.RequestID
	}
	if err := u.responseRepository.Insert(row); err != nil {
		return fmt.Errorf("failed to insert response record: %w", err)
	}
	for _, schedule := range resp.Schedules {
		if err := u.interpretSchedule(u.messageID, schedule); err != nil {
			return fmt.Errorf("failed to process Schedule %s: %w", schedule.RID, err)
		}
	}
	return nil
}
