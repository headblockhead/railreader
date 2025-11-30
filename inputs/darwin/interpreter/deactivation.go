package interpreter

import (
	"github.com/google/uuid"
	"github.com/headblockhead/railreader/inputs/darwin/unmarshaller"
	"github.com/jackc/pgx/v5"
)

// interpretDeactivation takes an unmarshalled Deactivation event, and adds it to the database.
func (u UnitOfWork) interpretDeactivation(deactivation unmarshaller.Deactivation) error {
	record, err := u.deactivationToRecord(deactivation)
	if err != nil {
		return err
	}
	err = u.insertDeactivation(record)
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

func (u UnitOfWork) deactivationToRecord(deactivation unmarshaller.Deactivation) (deactivationRecord, error) {
	var record deactivationRecord
	record.ID = uuid.New()
	record.MessageID = *u.messageID
	record.ScheduleID = deactivation.RID
	return record, nil
}

func (u UnitOfWork) insertDeactivation(record deactivationRecord) error {
	_, err := u.tx.Exec(u.ctx, `
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
	if err != nil {
		return err
	}
	return nil
}
