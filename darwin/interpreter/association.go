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
	mainScheduleLocations, err := u.scheduleLocationRepository.SelectManyByScheduleID(association.MainService.RID)
	if err != nil {
		return err
	}
	for _, loc := range mainScheduleLocations {
		if timesEqual(unmarshaller.LocationTimeIdentifiers{
			PublicArrivalTime:    loc.PublicArrivalTime,
			PublicDepartureTime:  loc.PublicDepartureTime,
			WorkingArrivalTime:   loc.WorkingArrivalTime,
			WorkingDepartureTime: loc.WorkingDepartureTime,
			WorkingPassingTime:   loc.WorkingPassingTime,
		}, association.MainService.LocationTimeIdentifiers) {
			row.MainScheduleLocationSequence = loc.Sequence
			break
		}
	}
	row.AssociatedScheduleID = association.AssociatedService.RID
	associatedScheduleLocations, err := u.scheduleLocationRepository.SelectManyByScheduleID(association.AssociatedService.RID)
	if err != nil {
		return err
	}
	for _, loc := range associatedScheduleLocations {
		if timesEqual(unmarshaller.LocationTimeIdentifiers{
			PublicArrivalTime:    loc.PublicArrivalTime,
			PublicDepartureTime:  loc.PublicDepartureTime,
			WorkingArrivalTime:   loc.WorkingArrivalTime,
			WorkingDepartureTime: loc.WorkingDepartureTime,
			WorkingPassingTime:   loc.WorkingPassingTime,
		}, association.AssociatedService.LocationTimeIdentifiers) {
			row.AssociatedScheduleLocationSequence = loc.Sequence
			break
		}
	}

	return u.associationRepository.Insert(row)
}
