package interpreter

import (
	"errors"
	"log/slog"
	"time"

	"github.com/headblockhead/railreader/darwin/unmarshaller"
	"github.com/jackc/pgx/v5"
)

func (u UnitOfWork) InterpretPushPortMessage(pport unmarshaller.PushPortMessage) error {
	u.log.Debug("interpreting a PushPortMessage")
	if pport.Version != unmarshaller.ExpectedPushPortVersion {
		// Warn, but attempt to continue anyway.
		u.log.Warn("PushPortMessage version does not match expected version", slog.String("expected_version", unmarshaller.ExpectedPushPortVersion), slog.String("actual_version", pport.Version))
	}

	alreadyExists, err := u.doesMessageRecordExist(*u.messageID)
	if err != nil {
		return err
	}

	if alreadyExists {
		err := u.updateMessageRecordTime(*u.messageID)
		if err != nil {
			return err
		}
	} else {
		record, err := u.messageToRecord(pport)
		if err != nil {
			return err
		}
		err = u.insertMessageRecord(record)
		if err != nil {
			return err
		}
	}

	/* if pport.NewFiles != nil {*/
	/*if err := u.handleNewFiles(pport.NewFiles); err != nil {*/
	/*return fmt.Errorf("failed to handle NewFiles: %w", err)*/
	/*}*/
	/*return nil*/
	/*}*/
	if pport.UpdateResponse != nil {
		if err := u.interpretResponse(pport.UpdateResponse); err != nil {
			return err
		}
		return nil
	}
	if pport.SnapshotResponse != nil {
		if err := u.interpretResponse(pport.SnapshotResponse); err != nil {
			return err
		}
		return nil
	}

	return errors.New("PushPortMessage was empty")
}

type messageRecord struct {
	ID              string
	Version         string
	SentAt          time.Time
	FirstReceivedAt time.Time
	LastReceivedAt  time.Time

	StatusCode        *string
	StatusDescription *string

	ResponseIsSnapshot  *bool
	RequestSource       *string
	RequestSourceSystem *string
	RequestID           *string
}

func (u UnitOfWork) doesMessageRecordExist(ID string) (bool, error) {
	row := u.tx.QueryRow(u.ctx, `SELECT id FROM darwin.messages WHERE id = @id;`, pgx.StrictNamedArgs{
		"id": ID,
	})
	var id string
	err := row.Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (u UnitOfWork) messageToRecord(message unmarshaller.PushPortMessage) (messageRecord, error) {
	var record messageRecord
	record.ID = *u.messageID
	record.Version = message.Version
	sentAt, err := time.ParseInLocation(time.RFC3339Nano, message.Timestamp, u.timezone)
	if err != nil {
		return record, err
	}
	record.SentAt = sentAt
	now := time.Now().In(u.timezone)
	record.FirstReceivedAt = now
	record.LastReceivedAt = now

	if message.StatusUpdate != nil {
		record.StatusCode = (*string)(&message.StatusUpdate.Code)
		record.StatusDescription = &message.StatusUpdate.Description
		record.RequestSourceSystem = message.StatusUpdate.SourceSystem
		record.RequestID = message.StatusUpdate.RequestID
	}
	if message.UpdateResponse != nil {
		isSnapshot := false
		record.ResponseIsSnapshot = &isSnapshot
		record.RequestSource = message.UpdateResponse.Source
		record.RequestSourceSystem = message.UpdateResponse.SourceSystem
		record.RequestID = message.UpdateResponse.RequestID
	}
	if message.SnapshotResponse != nil {
		isSnapshot := true
		record.ResponseIsSnapshot = &isSnapshot
		record.RequestSource = message.SnapshotResponse.Source
		record.RequestSourceSystem = message.SnapshotResponse.SourceSystem
		record.RequestID = message.SnapshotResponse.RequestID
	}
	return record, nil
}

func (u UnitOfWork) insertMessageRecord(record messageRecord) error {
	_, err := u.tx.Exec(u.ctx, `
	INSERT INTO darwin.messages (
		id
		,version
		,sent_at
		,first_received_at
		,last_received_at
		,status_code
		,status_description
		,response_is_snapshot
		,request_source
		,request_source_system
		,request_id
	) VALUES (
		@id
		,@version
		,@sent_at
		,@first_received_at
		,@last_received_at
		,@status_code
		,@status_description
		,@response_is_snapshot
		,@request_source
		,@request_source_system
		,@request_id
	);	`, pgx.StrictNamedArgs{
		"id":                    record.ID,
		"version":               record.Version,
		"sent_at":               record.SentAt,
		"first_received_at":     record.FirstReceivedAt,
		"last_received_at":      record.LastReceivedAt,
		"status_code":           record.StatusCode,
		"status_description":    record.StatusDescription,
		"response_is_snapshot":  record.ResponseIsSnapshot,
		"request_source":        record.RequestSource,
		"request_source_system": record.RequestSourceSystem,
		"request_id":            record.RequestID,
	})
	if err != nil {
		return err
	}
	return nil
}
func (u UnitOfWork) updateMessageRecordTime(ID string) error {
	_, err := u.tx.Exec(u.ctx, `
	UPDATE darwin.messages SET (
		last_received_at = @last_received_at
	) WHERE id = @id;
	`, pgx.StrictNamedArgs{
		"id":               ID,
		"last_received_at": time.Now().In(u.timezone),
	})
	if err != nil {
		return err
	}
	return nil
}

/*func (u UnitOfWork) handleNewFiles(tf *unmarshaller.NewFiles) error {*/
/*u.log.Debug("handling NewFiles")*/
/*// Filter for specific version numbers of the files we care about.*/
/*if strings.HasSuffix(tf.ReferenceFile, "_ref_v4.xml.gz") {*/
/*GetUnmarshalAndInterpretFile(u.log, u.fg, tf.ReferenceFile, unmarshaller.NewReference, u.InterpretReference)*/
/*}*/
/*if strings.HasSuffix(tf.TimetableFile, "_v8.xml.gz") {*/
/*GetUnmarshalAndInterpretFile(u.log, u.fg, tf.TimetableFile, unmarshaller.NewTimetable, u.InterpretTimetable)*/
/*}*/
/*return nil*/
/*}*/

/*func (u UnitOfWork) InterpretFromPath(path string) error {*/
/*u.log.Debug("handling a filename", slog.String("path", path))*/
/*if strings.HasSuffix(path, "_ref_v4.xml.gz") {*/
/*return GetUnmarshalAndInterpretFile(u.log, u.fg, path, unmarshaller.NewReference, u.InterpretReference)*/
/*}*/
/*if strings.HasSuffix(path, "_v8.xml.gz") {*/
/*return GetUnmarshalAndInterpretFile(u.log, u.fg, path, unmarshaller.NewTimetable, u.InterpretTimetable)*/
/*}*/
/*u.log.Info("filename does not match any known patterns, ignoring", slog.String("path", path))*/
/*return nil*/
/*}*/

/*func GetUnmarshalAndInterpretFile[T any](log *slog.Logger, fg filegetter.FileGetter, path string, unmarshal func(string) (T, error), interpret func(T, string) error) error {*/
/*log.Debug("getting file", slog.String("path", path))*/
/*file, err := fg.Get(path)*/
/*if err != nil {*/
/*return fmt.Errorf("failed to get from filegetter: %w", err)*/
/*}*/
/*log.Debug("file gotten")*/
/*reader, err := gzip.NewReader(file)*/
/*if err != nil {*/
/*return fmt.Errorf("failed to create gzip reader: %w", err)*/
/*}*/
/*defer reader.Close()*/
/*contents, err := io.ReadAll(reader)*/
/*if err != nil {*/
/*return fmt.Errorf("failed to read all of gzip reader: %w", err)*/
/*}*/
/*log.Debug("file read")*/
/*data, err := unmarshal(string(contents))*/
/*if err != nil {*/
/*return err*/
/*}*/
/*log.Debug("file unmarshalled")*/
/*if err := interpret(data, path); err != nil {*/
/*return err*/
/*}*/
/*log.Debug("file interpreted")*/
/*return nil*/
/*}*/

func (u UnitOfWork) interpretResponse(resp *unmarshaller.Response) error {
	for _, alarm := range resp.Alarms {
		if err := u.interpretAlarm(alarm); err != nil {
			return err
		}
	}
	for _, association := range resp.Associations {
		if err := u.interpretAssociation(association); err != nil {
			return err
		}
	}
	for _, deactivation := range resp.Deactivations {
		if err := u.interpretDeactivation(deactivation); err != nil {
			return err
		}
	}
	for _, forecastTime := range resp.ForecastTimes {
		if err := u.interpretForecast(forecastTime); err != nil {
			return err
		}
	}
	for _, formationLoading := range resp.FormationLoadings {
		if err := u.interpretFormationLoading(formationLoading); err != nil {
			return err
		}
	}
	for _, formation := range resp.Formations {
		if err := u.interpretFormation(formation); err != nil {
			return err
		}
	}
	for _, headcodeChange := range resp.HeadcodeChanges {
		if err := u.interpretHeadcodeChange(headcodeChange); err != nil {
			return err
		}
	}
	for _, schedule := range resp.Schedules {
		if err := u.interpretSchedule(schedule); err != nil {
			return err
		}
	}
	for _, serviceLoading := range resp.ServiceLoadings {
		if err := u.interpretServiceLoading(serviceLoading); err != nil {
			return err
		}
	}
	for _, stationMessage := range resp.StationMessages {
		if err := u.interpretStationMessage(stationMessage); err != nil {
			return err
		}
	}
	/* for _, trainAlert := range resp.TrainAlerts {*/
	/*if err := u.interpretTrainAlert(trainAlert); err != nil {*/
	/*return err*/
	/*}*/
	/*}*/
	//for _,trainOrder:=range resp.TrainOrders{if err:=u.interpretTrainOrder(trainOrder);err!=nil{return err}}
	return nil
}
