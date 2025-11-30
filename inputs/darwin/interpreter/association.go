package interpreter

import (
	"github.com/google/uuid"
	"github.com/headblockhead/railreader/inputs/darwin/unmarshaller"
	"github.com/jackc/pgx/v5"
)

// interpretAssociation takes an unmarshalled Association object, converts it into a database record, and inserts it into the database.
func (u UnitOfWork) interpretAssociation(association unmarshaller.Association) error {
	record, err := u.associationToRecord(association)
	if err != nil {
		return err
	}
	err = u.insertAssociationRecord(record)
	if err != nil {
		return err
	}
	return nil
}

type associationRecord struct {
	ID uuid.UUID

	MessageID   *string
	TimetableID *string

	LocationID  string
	Category    string
	IsCancelled bool
	IsDeleted   bool

	MainScheduleID                   string
	MainScheduleWorkingArrivalTime   *string
	MainScheduleWorkingPassingTime   *string
	MainScheduleWorkingDepartureTime *string
	MainSchedulePublicArrivalTime    *string
	MainSchedulePublicDepartureTime  *string

	AssociatedScheduleID                   string
	AssociatedScheduleWorkingArrivalTime   *string
	AssociatedScheduleWorkingPassingTime   *string
	AssociatedScheduleWorkingDepartureTime *string
	AssociatedSchedulePublicArrivalTime    *string
	AssociatedSchedulePublicDepartureTime  *string
}

// associationToRecord converts an unmarshalled Association object into an association database record.
func (u UnitOfWork) associationToRecord(association unmarshaller.Association) (associationRecord, error) {
	var record associationRecord
	record.ID = uuid.New()
	record.MessageID = u.messageID
	record.TimetableID = u.timetableID
	record.LocationID = association.TIPLOC
	record.Category = string(association.Category)
	record.IsCancelled = association.Cancelled
	record.IsDeleted = association.Deleted
	record.MainScheduleID = association.MainService.RID
	record.MainScheduleWorkingArrivalTime = association.MainService.WorkingArrivalTime
	record.MainScheduleWorkingPassingTime = association.MainService.WorkingPassingTime
	record.MainScheduleWorkingDepartureTime = association.MainService.WorkingDepartureTime
	record.MainSchedulePublicArrivalTime = association.MainService.PublicArrivalTime
	record.MainSchedulePublicDepartureTime = association.MainService.PublicDepartureTime
	record.AssociatedScheduleID = association.AssociatedService.RID
	record.AssociatedScheduleWorkingArrivalTime = association.AssociatedService.WorkingArrivalTime
	record.AssociatedScheduleWorkingPassingTime = association.AssociatedService.WorkingPassingTime
	record.AssociatedScheduleWorkingDepartureTime = association.AssociatedService.WorkingDepartureTime
	record.AssociatedSchedulePublicArrivalTime = association.AssociatedService.PublicArrivalTime
	record.AssociatedSchedulePublicDepartureTime = association.AssociatedService.PublicDepartureTime
	return record, nil
}

// insertAssociationRecord inserts a new association record in the database.
func (u UnitOfWork) insertAssociationRecord(record associationRecord) error {
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
			,main_schedule_working_arrival_time
			,main_schedule_working_passing_time
			,main_schedule_working_departure_time
			,main_schedule_public_arrival_time
			,main_schedule_public_departure_time
			,associated_schedule_id
			,associated_schedule_working_arrival_time
			,associated_schedule_working_passing_time
			,associated_schedule_working_departure_time
			,associated_schedule_public_arrival_time
			,associated_schedule_public_departure_time
		)	VALUES (
			@id
			,@message_id
			,@timetable_id
			,@location_id
			,@category
			,@is_cancelled
			,@is_deleted
			,@main_schedule_id
			,@main_schedule_working_arrival_time
			,@main_schedule_working_passing_time
			,@main_schedule_working_departure_time
			,@main_schedule_public_arrival_time
			,@main_schedule_public_departure_time
			,@associated_schedule_id
			,@associated_schedule_working_arrival_time
			,@associated_schedule_working_passing_time
			,@associated_schedule_working_departure_time
			,@associated_schedule_public_arrival_time
			,@associated_schedule_public_departure_time
		);
	`, pgx.StrictNamedArgs{
		"id":                                         record.ID,
		"message_id":                                 record.MessageID,
		"timetable_id":                               record.TimetableID,
		"location_id":                                record.LocationID,
		"category":                                   record.Category,
		"is_cancelled":                               record.IsCancelled,
		"is_deleted":                                 record.IsDeleted,
		"main_schedule_id":                           record.MainScheduleID,
		"main_schedule_working_arrival_time":         record.MainScheduleWorkingArrivalTime,
		"main_schedule_working_passing_time":         record.MainScheduleWorkingPassingTime,
		"main_schedule_working_departure_time":       record.MainScheduleWorkingDepartureTime,
		"main_schedule_public_arrival_time":          record.MainSchedulePublicArrivalTime,
		"main_schedule_public_departure_time":        record.MainSchedulePublicDepartureTime,
		"associated_schedule_id":                     record.AssociatedScheduleID,
		"associated_schedule_working_arrival_time":   record.AssociatedScheduleWorkingArrivalTime,
		"associated_schedule_working_passing_time":   record.AssociatedScheduleWorkingPassingTime,
		"associated_schedule_working_departure_time": record.AssociatedScheduleWorkingDepartureTime,
		"associated_schedule_public_arrival_time":    record.AssociatedSchedulePublicArrivalTime,
		"associated_schedule_public_departure_time":  record.AssociatedSchedulePublicDepartureTime,
	})
	if err != nil {
		return err
	}
	return nil
}
