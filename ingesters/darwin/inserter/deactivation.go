package inserter

import (
	"github.com/google/uuid"
	"github.com/headblockhead/railreader/ingesters/darwin/unmarshaller"
	"github.com/jackc/pgx/v5"
)

func (u *UnitOfWork) insertDeactivation(deactivation unmarshaller.Deactivation) error {
	record, err := u.deactivationToRecord(deactivation)
	if err != nil {
		return err
	}
	err = u.insertDeactivationRecord(record)
	if err != nil {
		return err
	}
	return nil
}

type deactivationRecord struct {
	ID         uuid.UUID
	MessageID  string
	ScheduleID string
}

func (u *UnitOfWork) deactivationToRecord(deactivation unmarshaller.Deactivation) (deactivationRecord, error) {
	var record deactivationRecord
	record.ID = uuid.New()
	record.MessageID = *u.messageID
	record.ScheduleID = deactivation.RID
	return record, nil
}

func (u *UnitOfWork) insertDeactivationRecord(record deactivationRecord) error {
	u.batch.Queue(`
		INSERT INTO darwin.deactivations (
			id
			,message_id
			,schedule_id
		) VALUES (
			@id
			,@message_id
			,@schedule_id
		);
	`, pgx.StrictNamedArgs{
		"id":          record.ID,
		"message_id":  u.messageID,
		"schedule_id": record.ScheduleID,
	})
	return nil
}
