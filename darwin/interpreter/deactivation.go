package interpreter

import (
	"github.com/headblockhead/railreader/darwin/unmarshaller"
	"github.com/jackc/pgx/v5"
)

// interpretDeactivation takes an unmarshalled Deactivation event, and adds it to the database.
func (u UnitOfWork) interpretDeactivation(deactivation unmarshaller.Deactivation) error {
	_, err := u.tx.Exec(u.ctx, `
		INSERT INTO darwin.deactivations (
			id
			,schedule_id
			,message_id
		) VALUES (
			@id
			,@schedule_id
			,@message_id
		);
	`, pgx.StrictNamedArgs{
		"schedule_id": deactivation.RID,
		"message_id":  u.messageID,
	})
	if err != nil {
		return err
	}
	return nil
}
