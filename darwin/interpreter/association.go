package interpreter

import (
	"github.com/headblockhead/railreader/darwin/repository"
	"github.com/headblockhead/railreader/darwin/unmarshaller"
)

func (u UnitOfWork) interpretAssociation(association unmarshaller.Association) error {
	var row repository.AssociationRow
	row.Category = string(association.Category)
	row.IsCancelled = association.Cancelled
	row.IsDeleted = association.Deleted
	row.MainScheduleID = association.MainService.RID
	row.AssociatedScheduleID = association.AssociatedService.RID
	// TODO: sequence index
	/* mainLocation, err := u.locationFromTIPLOCAndTimes(association.MainService.RID, association.TIPLOC, association.MainService.LocationTimeIdentifiers)*/
	/*if err != nil {*/
	/*return err*/
	/*}*/
	/*row.MainScheduleLocationSequence = mainLocation.Sequence*/
	/*associatedLocation, err := u.locationFromTIPLOCAndTimes(association.AssociatedService.RID, association.TIPLOC, association.AssociatedService.LocationTimeIdentifiers)*/
	/*if err != nil {*/
	/*return err*/
	/*}*/
	/*row.AssociatedScheduleLocationSequence = associatedLocation.Sequence*/

	return u.associationRepository.Insert(row)
}
