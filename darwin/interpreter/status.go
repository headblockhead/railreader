package interpreter

import (
	"github.com/headblockhead/railreader/darwin/repository"
	"github.com/headblockhead/railreader/darwin/unmarshaller"
)

func (u UnitOfWork) interpretStatus(status unmarshaller.Status) error {
	u.log.Debug("interpreting a Status")
	// TODO: create status repository
	var row repository.StatusRow
	row.SourceSystem = status.SourceSystem
	row.RequestID = status.RequestID
	row.Code = string(status.Code)
	row.Description = status.Description
	if err := u.statusRepository.Insert(row); err != nil {
		return err
	}
	return nil
}
