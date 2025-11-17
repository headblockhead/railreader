package interpreter

import (
	"errors"

	"github.com/google/uuid"
	"github.com/headblockhead/railreader/darwin/unmarshaller"
	"github.com/jackc/pgx/v5"
)

// interpretAlarm takes an unmarshalled Alarm event, and records it in the database.
func (u UnitOfWork) interpretAlarm(alarm unmarshaller.Alarm) error {
	record, err := u.alarmToRecord(alarm)
	if err != nil {
		return err
	}
	err = u.insertAlarmRecord(record)
	if err != nil {
		return err
	}
	return nil
}

type alarmRecord struct {
	ID uuid.UUID

	MessageID *string

	AlarmID    int
	HasCleared bool

	TrainDescriberFailure    *string
	AllTrainDescribersFailed *bool
	TyrellFailed             *bool
}

// alarmToRecord converts an unmarshalled Alarm object into an alarm database record.
func (u UnitOfWork) alarmToRecord(alarm unmarshaller.Alarm) (alarmRecord, error) {
	var record alarmRecord
	record.ID = uuid.New()
	record.MessageID = u.messageID
	if alarm.ClearedAlarm != nil {
		record.HasCleared = true
		record.AlarmID = *alarm.ClearedAlarm
	} else if alarm.NewAlarm != nil {
		record.TrainDescriberFailure = alarm.NewAlarm.TDFailure
		record.AllTrainDescribersFailed = (*bool)(&alarm.NewAlarm.TDTotalFailure)
		record.TyrellFailed = (*bool)(&alarm.NewAlarm.TyrellTotalFailure)
	} else {
		return record, errors.New("no alarm data present")
	}
	return record, nil
}

// insertAlarmRecord inserts a new alarm record in the database.
func (u UnitOfWork) insertAlarmRecord(record alarmRecord) error {
	_, err := u.tx.Exec(u.ctx, `
		INSERT INTO darwin.alarms (
			id
			,message_id
			,alarm_id
			,has_cleared
			,train_describer_failure
			,all_train_describers_failed
			,tyrell_failed
		)	VALUES (
			@id
			,@message_id
			,@alarm_id
			,@has_cleared
			,@train_describer_failure
			,@all_train_describers_failed
			,@tyrell_failed
		);
	`, pgx.StrictNamedArgs{
		"id":                          record.ID,
		"message_id":                  record.MessageID,
		"alarm_id":                    record.AlarmID,
		"has_cleared":                 record.HasCleared,
		"train_describer_failure":     record.TrainDescriberFailure,
		"all_train_describers_failed": record.AllTrainDescribersFailed,
		"tyrell_failed":               record.TyrellFailed,
	})
	if err != nil {
		return err
	}
	return nil
}
