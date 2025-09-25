package interpreter

import (
	"github.com/headblockhead/railreader/darwin/repository"
	"github.com/headblockhead/railreader/darwin/unmarshaller"
)

func (u *UnitOfWork) interpretAssociation(association unmarshaller.Association) error {
	var row repository.AssociationRow
	row.Category = string(association.Category)
	row.IsCancelled = association.Cancelled
	row.IsDeleted = association.Deleted
	row.MainScheduleID = association.MainService.RID
	row.AssociatedScheduleID = association.AssociatedService.RID

	// TODO: link TIPLOC+times to a schedule sequence number

	return u.associationRepository.Insert(row)
}
