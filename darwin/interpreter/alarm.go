package interpreter

import (
	"errors"

	"github.com/headblockhead/railreader/darwin/unmarshaller"
	"github.com/jackc/pgx/v5"
)

// interpretAlarm takes an unmarshalled Alarm event, and records it in the database.
func (u UnitOfWork) interpretAlarm(alarm unmarshaller.Alarm) error {
	if alarm.ClearedAlarm != nil {
		if _, err := u.tx.Exec(u.ctx, `
			INSERT INTO darwin.alarms_cleared (
				id
				,message_id
			) VALUES (
				@alarm_id
				,@message_id
			);`, pgx.StrictNamedArgs{
			"alarm_id":   alarm.ClearedAlarm,
			"message_id": u.messageID,
		}); err != nil {
			return err
		}
	} else if alarm.NewAlarm != nil {
		record, err := u.newAlarmToRecord(*alarm.NewAlarm)
		if err != nil {
			return err
		}
		err = u.insertOneAlarmRecord(record)
		if err != nil {
			return err
		}
	} else {
		return errors.New("no alarm data present")
	}
	return nil
}

type alarmRecord struct {
	ID                       int
	MessageID                *string
	TrainDescriberFailure    *string
	AllTrainDescribersFailed *bool
	TyrellFailed             *bool
}

// newAlarmToRecord converts an unmarshalled NewAlarm object into an alarm database record.
func (u UnitOfWork) newAlarmToRecord(alarm unmarshaller.NewAlarm) (alarmRecord, error) {
	var record alarmRecord
	record.ID = alarm.ID
	record.MessageID = u.messageID
	record.TrainDescriberFailure = alarm.TDFailure
	record.AllTrainDescribersFailed = (*bool)(&alarm.TDTotalFailure)
	record.TyrellFailed = (*bool)(&alarm.TyrellTotalFailure)
	return record, nil
}

// insertOneAlarmRecord inserts a new alarm record in the database, ignoring conflicts.
func (u UnitOfWork) insertOneAlarmRecord(record alarmRecord) error {
	_, err := u.tx.Exec(u.ctx, `
		INSERT INTO darwin.alarms (
			id
			,message_id
			,train_describer_failure
			,all_train_describers_failed
			,tyrell_failed
		)	VALUES (
			@alarm_id
			,@message_id
			,@train_describer_failure
			,@all_train_describers_failed
			,@tyrell_failed
		) ON CONFLICT DO NOTHING;
		`, pgx.StrictNamedArgs{
		"id":                          record.ID,
		"message_id":                  record.MessageID,
		"train_describer_failure":     record.TrainDescriberFailure,
		"all_train_describers_failed": record.AllTrainDescribersFailed,
		"tyrell_failed":               record.TyrellFailed,
	})
	if err != nil {
		return err
	}
	return nil
}
