package interpreter

import (
	"github.com/headblockhead/railreader/darwin/repository"
	"github.com/headblockhead/railreader/darwin/unmarshaller"
)

func (u UnitOfWork) interpretAssociation(association unmarshaller.Association) error {
	row, err := u.associationToRow(association)
	if err != nil {
		return err
	}
	return u.associationRepository.Insert(row)
}

func (u UnitOfWork) associationToRow(association unmarshaller.Association) (row repository.AssociationRow, err error) {
	row.MessageID = u.messageID
	row.TimetableID = u.timetableID
	row.Category = string(association.Category)
	row.IsCancelled = association.Cancelled
	row.IsDeleted = association.Deleted
	row.MainScheduleID = association.MainService.RID
	mainScheduleLocations, err := u.scheduleLocationRepository.SelectManyByScheduleID(association.MainService.RID)
	if err != nil {
		return row, err
	}
	for _, loc := range mainScheduleLocations {
		if timesEqual(unmarshaller.LocationTimeIdentifiers{
			PublicArrivalTime:    loc.PublicArrivalTime,
			PublicDepartureTime:  loc.PublicDepartureTime,
			WorkingArrivalTime:   loc.WorkingArrivalTime,
			WorkingDepartureTime: loc.WorkingDepartureTime,
			WorkingPassingTime:   loc.WorkingPassingTime,
		}, association.MainService.LocationTimeIdentifiers) && loc.LocationID == association.TIPLOC {
			row.MainScheduleLocationSequence = loc.Sequence
			break
		}
	}
	row.AssociatedScheduleID = association.AssociatedService.RID
	associatedScheduleLocations, err := u.scheduleLocationRepository.SelectManyByScheduleID(association.AssociatedService.RID)
	if err != nil {
		return row, err
	}
	for _, loc := range associatedScheduleLocations {
		if timesEqual(unmarshaller.LocationTimeIdentifiers{
			PublicArrivalTime:    loc.PublicArrivalTime,
			PublicDepartureTime:  loc.PublicDepartureTime,
			WorkingArrivalTime:   loc.WorkingArrivalTime,
			WorkingDepartureTime: loc.WorkingDepartureTime,
			WorkingPassingTime:   loc.WorkingPassingTime,
		}, association.AssociatedService.LocationTimeIdentifiers) && loc.LocationID == association.TIPLOC {
			row.AssociatedScheduleLocationSequence = loc.Sequence
			break
		}
	}

	return row, nil
}
