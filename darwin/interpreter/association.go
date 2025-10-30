package interpreter

import (
	"github.com/google/uuid"
	"github.com/headblockhead/railreader/darwin/unmarshaller"
)

// interpretAssociation takes an unmarshalled Association object, converts it into a database record, and inserts it into the database.
func (u UnitOfWork) interpretAssociation(association unmarshaller.Association) error {
	record, err := u.associationToRecord(association)
	if err != nil {
		return err
	}
	return u.insertOneAssociationRecord(record)
}

type AssociationRecord struct {
	AssociationID                      uuid.UUID
	MessageID                          *string
	TimetableID                        *string
	Category                           string
	IsCancelled                        bool
	IsDeleted                          bool
	MainScheduleID                     string
	MainScheduleLocationSequence       int
	AssociatedScheduleID               string
	AssociatedScheduleLocationSequence int
}

// associationToRecord converts an unmarshalled Association object into an association database record.
// It requires both of the schedules (and their locations) to already exist in the database.
func (u UnitOfWork) associationToRecord(association unmarshaller.Association) (AssociationRecord, error) {
	var record AssociationRecord
	record.AssociationID = uuid.New()
	record.MessageID = u.messageID
	record.TimetableID = u.timetableID
	record.Category = string(association.Category)
	record.IsCancelled = association.Cancelled
	record.IsDeleted = association.Deleted
	record.MainScheduleID = association.MainService.RID
	var err error
	record.MainScheduleLocationSequence, err = u.findLocationSequence(association.MainService.RID, association.TIPLOC, association.MainService.LocationTimeIdentifiers)
	if err != nil {
		return record, err
	}
	record.AssociatedScheduleID = association.AssociatedService.RID
	record.AssociatedScheduleLocationSequence, err = u.findLocationSequence(association.AssociatedService.RID, association.TIPLOC, association.AssociatedService.LocationTimeIdentifiers)
	if err != nil {
		return record, err
	}
	return record, nil
}

// insertOneAssociationRecord inserts a single association record into the database.
func (u UnitOfWork) insertOneAssociationRecord(record AssociationRecord) error {
}
