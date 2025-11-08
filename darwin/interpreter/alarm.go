package interpreter

import (
	"errors"
	"time"

	"github.com/headblockhead/railreader/darwin/unmarshaller"
	"github.com/jackc/pgx/v5"
)

// interpretAlarm takes an unmarshalled Alarm event, and records it in the database.
func (u UnitOfWork) interpretAlarm(alarm unmarshaller.Alarm) error {
	if alarm.ClearedAlarm != nil {
		_, err := u.tx.Exec(u.ctx, `UPDATE darwin.alarms SET has_cleared = @has_cleared, cleared_at = @cleared_at WHERE id = @alarm_id;`, pgx.StrictNamedArgs{
			"has_cleared": true,
			"cleared_at":  time.Now(),
			"alarm_id":    *alarm.ClearedAlarm,
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				u.log.Warn("tried to clear non-existent alarm", "alarm_id", *alarm.ClearedAlarm)
				return nil
			}
			return err
		}
	} else if alarm.NewAlarm != nil {
		record, err := u.newAlarmToRecord(*alarm.NewAlarm)
		if err != nil {
			return err
		}
		err = u.upsertOneAlarmRecord(record)
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
	ReceivedAt               time.Time
	HasCleared               bool
	ClearedAt                *time.Time
	TrainDescriberFailure    *string
	AllTrainDescribersFailed *bool
	TyrellFailed             *bool
}

// newAlarmToRecord converts an unmarshalled NewAlarm object into an alarm database record.
func (u UnitOfWork) newAlarmToRecord(alarm unmarshaller.NewAlarm) (alarmRecord, error) {
	var record alarmRecord
	record.ID = alarm.ID
	record.MessageID = u.messageID
	record.ReceivedAt = time.Now()
	record.HasCleared = false
	record.ClearedAt = nil
	record.TrainDescriberFailure = alarm.TDFailure
	record.AllTrainDescribersFailed = (*bool)(&alarm.TDTotalFailure)
	record.TyrellFailed = (*bool)(&alarm.TyrellTotalFailure)
	return record, nil
}

// upsertOneAlarmRecord inserts a new (or updates an existing) alarm record in the database.
func (u UnitOfWork) upsertOneAlarmRecord(record alarmRecord) error {
	_, err := u.tx.Exec(u.ctx, `
		INSERT INTO darwin.alarms (
			id
			,message_id
			,received_at
			,has_cleared
			,cleared_at
			,train_describer_failure
			,all_train_describers_failed
			,tyrell_failed
		)	VALUES (
			@id
			,@message_id
			,@received_at
			,@has_cleared
			,@cleared_at
			,@train_describer_failure
			,@all_train_describers_failed
			,@tyrell_failed
		) ON CONFLICT (id) UPDATE SET (
			message_id = EXCLUDED.message_id
			,received_at = EXCLUDED.received_at
			,has_cleared = EXCLUDED.has_cleared
			,cleared_at = EXCLUDED.cleared_at
			,train_describer_failure = EXCLUDED.train_describer_failure
			,all_train_describers_failed = EXCLUDED.all_train_describers_failed
			,tyrell_failed = EXCLUDED.tyrell_failed
		);
		`, pgx.StrictNamedArgs{
		"id":                          record.ID,
		"message_id":                  record.MessageID,
		"received_at":                 record.ReceivedAt,
		"has_cleared":                 record.HasCleared,
		"cleared_at":                  record.ClearedAt,
		"train_describer_failure":     record.TrainDescriberFailure,
		"all_train_describers_failed": record.AllTrainDescribersFailed,
		"tyrell_failed":               record.TyrellFailed,
	})
	if err != nil {
		return err
	}
	return nil
}
