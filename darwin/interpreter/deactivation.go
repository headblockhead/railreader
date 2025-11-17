package interpreter

import (
	"github.com/google/uuid"
	"github.com/headblockhead/railreader/darwin/unmarshaller"
	"github.com/jackc/pgx/v5"
)

// interpretDeactivation takes an unmarshalled Deactivation event, and adds it to the database.
func (u UnitOfWork) interpretDeactivation(deactivation unmarshaller.Deactivation) error {
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
		"message_id":  u.messageID,
		"schedule_id": deactivation.RID,
	})
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

func (u UnitOfWork) insertDeactivation(record deactivationRecord) error {
}
