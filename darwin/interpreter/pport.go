package interpreter

import (
	"compress/gzip"
	"errors"
	"io"
	"io/fs"
	"log/slog"
	"strings"
	"time"

	"github.com/headblockhead/railreader/darwin/unmarshaller"
	"github.com/jackc/pgx/v5"
)

func (u *UnitOfWork) InterpretPushPortMessage(pport unmarshaller.PushPortMessage) error {
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

	if pport.NewFiles != nil {
		if err := u.handleNewFiles(pport.NewFiles); err != nil {
			return err
		}
		return nil
	}
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

	return nil
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

func (u *UnitOfWork) doesMessageRecordExist(ID string) (bool, error) {
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

func (u *UnitOfWork) messageToRecord(message unmarshaller.PushPortMessage) (messageRecord, error) {
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

func (u *UnitOfWork) insertMessageRecord(record messageRecord) error {
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

func (u *UnitOfWork) updateMessageRecordTime(ID string) error {
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

func (u *UnitOfWork) handleNewFiles(tf *unmarshaller.NewFiles) error {
	if strings.HasSuffix(tf.ReferenceFile, unmarshaller.ExpectedReferenceFileSuffix) {
		file, err := u.fs.Open("PPTimetable/" + tf.ReferenceFile)
		if err != nil {
			return err
		}
		return u.InterpretReferenceFile(file)
	}
	if strings.HasSuffix(tf.TimetableFile, unmarshaller.ExpectedTimetableFileSuffix) {
		//file, err := u.fs.Open("PPTimetable/" + tf.TimetableFile)
		//if err != nil {
		//return err
		//}
		//return u.InterpretTimetableFile(file)
	}
	return nil
}

func (u *UnitOfWork) InterpretReferenceFile(file fs.File) error {
	bytes, err := decompressAndReadGzipFile(file)
	if err != nil {
		return err
	}
	reference, err := unmarshaller.NewReference(string(bytes))
	if err != nil {
		return err
	}
	err = u.InterpretReference(reference)
	if err != nil {
		return err
	}
	return nil
}

// func (u *UnitOfWork) InterpretTimetableFile(file fs.File) error {
// bytes, err := decompressAndReadGzipFile(file)
// if err != nil {
// return err
// }
// timetable, err := unmarshaller.NewTimetable(string(bytes))
// if err != nil {
// return err
// }
// err = u.InterpretTimetable(timetable)
// if err != nil {
// return err
// }
// return nil
// }

func decompressAndReadGzipFile(file fs.File) ([]byte, error) {
	reader, err := gzip.NewReader(file)
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	contents, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	return contents, nil
}

func (u *UnitOfWork) interpretResponse(resp *unmarshaller.Response) error {
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
	for _, trainAlert := range resp.TrainAlerts {
		if err := u.interpretTrainAlert(trainAlert); err != nil {
			return err
		}
	}
	//for _,trainOrder:=range resp.TrainOrders{if err:=u.interpretTrainOrder(trainOrder);err!=nil{return err}}
	return nil
}
