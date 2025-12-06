package inserter

import (
	"errors"

	"github.com/google/uuid"
	"github.com/headblockhead/railreader/ingesters/darwin/unmarshaller"
	"github.com/jackc/pgx/v5"
)

func (u *UnitOfWork) insertAlarm(alarm unmarshaller.Alarm) error {
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

func (u *UnitOfWork) alarmToRecord(alarm unmarshaller.Alarm) (alarmRecord, error) {
	var record alarmRecord
	record.ID = uuid.New()
	record.MessageID = u.messageID
	if alarm.ClearedAlarm != nil {
		record.HasCleared = true
		record.AlarmID = *alarm.ClearedAlarm
		return record, nil
	}
	if alarm.NewAlarm != nil {
		record.TrainDescriberFailure = alarm.NewAlarm.TDFailure
		record.AllTrainDescribersFailed = (*bool)(&alarm.NewAlarm.TDTotalFailure)
		record.TyrellFailed = (*bool)(&alarm.NewAlarm.TyrellTotalFailure)
		return record, nil
	}
	return record, errors.New("no alarm data present")
}

func (u *UnitOfWork) insertAlarmRecord(record alarmRecord) error {
	u.batch.Queue(`
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
	return nil
}
