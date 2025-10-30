package interpreter

import (
	"github.com/headblockhead/railreader/darwin/unmarshaller"
	"github.com/jackc/pgx/v5"
)

func (u UnitOfWork) interpretDeactivation(deactivation unmarshaller.Deactivation) error {
	_, err := u.tx.Exec(u.ctx, `UPDATE darwin.schedules SET is_active = FALSE WHERE id = @schedule_id`, pgx.StrictNamedArgs{"schedule_id": deactivation.RID})
	if err != nil {
		return err
	}
	return nil
}
