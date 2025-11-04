package interpreter

import (
	"github.com/google/uuid"
	"github.com/headblockhead/railreader/darwin/unmarshaller"
	"github.com/jackc/pgx/v5"
)

// interpretAssociation takes an unmarshalled Association object, converts it into a database record, and inserts it into the database.
func (u UnitOfWork) interpretAssociation(association unmarshaller.Association) error {
	record, err := u.associationToRecord(association)
	if err != nil {
		return err
	}
	return u.upsertOneAssociationRecord(record)
}

type AssociationRecord struct {
	ID                                 uuid.UUID
	MessageID                          *string
	TimetableID                        *string
	LocationID                         string
	Category                           string
	IsCancelled                        *bool
	IsDeleted                          *bool
	MainScheduleID                     string
	MainScheduleLocationSequence       int
	AssociatedScheduleID               string
	AssociatedScheduleLocationSequence int
}

// associationToRecord converts an unmarshalled Association object into an association database record.
// It requires both of the schedules (and their locations) to already exist in the database.
func (u UnitOfWork) associationToRecord(association unmarshaller.Association) (AssociationRecord, error) {
	var record AssociationRecord
	record.ID = uuid.New()
	record.MessageID = u.messageID
	record.TimetableID = u.timetableID
	record.LocationID = association.TIPLOC
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

// upsertOneAssociationRecord inserts a new (or updates an existing) association record in the database.
func (u UnitOfWork) upsertOneAssociationRecord(record AssociationRecord) error {
	_, err := u.tx.Exec(u.ctx, `
		INSERT INTO darwin.associations (
			id
			,message_id
			,timetable_id
			,location_id
			,category
			,is_cancelled
			,is_deleted
			,main_schedule_id
			,main_schedule_location_sequence
			,associated_schedule_id
			,associated_schedule_location_sequence
		)	VALUES (
			@id
			,@message_id
			,@timetable_id
			,@location_id
			,@category
			,@is_cancelled
			,@is_deleted
			,@main_schedule_id
			,@main_schedule_location_sequence
			,@associated_schedule_id
			,@associated_schedule_location_sequence
		) ON CONFLICT (id) DO UPDATE SET (
			message_id = EXCLUDED.message_id
			,timetable_id = EXCLUDED.timetable_id
			,location_id = EXCLUDED.location_id
			,category = EXCLUDED.category
			,is_cancelled = EXCLUDED.is_cancelled
			,is_deleted = EXCLUDED.is_deleted
			,main_schedule_id = EXCLUDED.main_schedule_id
			,main_schedule_location_sequence = EXCLUDED.main_schedule_location_sequence
			,associated_schedule_id = EXCLUDED.associated_schedule_id
			,associated_schedule_location_sequence = EXCLUDED.associated_schedule_location_sequence
		);
		`, pgx.StrictNamedArgs{
		"id":                                    record.ID,
		"message_id":                            record.MessageID,
		"timetable_id":                          record.TimetableID,
		"location_id":                           record.LocationID,
		"category":                              record.Category,
		"is_cancelled":                          record.IsCancelled,
		"is_deleted":                            record.IsDeleted,
		"main_schedule_id":                      record.MainScheduleID,
		"main_schedule_location_sequence":       record.MainScheduleLocationSequence,
		"associated_schedule_id":                record.AssociatedScheduleID,
		"associated_schedule_location_sequence": record.AssociatedScheduleLocationSequence,
	})
	if err != nil {
		return err
	}
	return nil
}
