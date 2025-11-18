package interpreter

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/headblockhead/railreader/darwin/unmarshaller"
	"github.com/jackc/pgx/v5"
)

func (u UnitOfWork) interpretSchedule(schedule unmarshaller.Schedule) error {
	scheduleRecord, scheduleLocationRecords, err := u.scheduleToRecords(schedule)
	if err != nil {
		return err
	}
	err = u.insertScheduleRecord(scheduleRecord)
	if err != nil {
		return err
	}
	err = u.insertScheduleLocationRecords(scheduleLocationRecords)
	if err != nil {
		return err
	}
	return nil
}

type scheduleRecord struct {
	ID uuid.UUID

	MessageID   *string
	TimetableID *string

	ScheduleID         string
	UID                string
	ScheduledStartDate string

	Headcode           string
	RetailServiceID    string
	TOCid              string
	Service            string
	Category           string
	IsPassengerService bool
	IsActive           bool
	IsDeleted          bool
	IsCharter          bool

	CancellationReasonID             *int
	CancellationReasonLocationID     *string
	CancellationReasonIsNearLocation *bool

	DivertedViaLocationID *string

	DiversionReasonID             *int
	DiversionReasonLocationID     *string
	DiversionReasonIsNearLocation *bool

	IsCancelled bool
}

type scheduleLocationRecord struct {
	ID uuid.UUID

	ScheduleUUID uuid.UUID
	ScheduleID   string

	LocationID            string
	Activities            []string
	PlannedActivities     []string
	IsCancelled           bool
	FormationID           *string
	IsAffectedByDiversion bool

	Type                       string
	WorkingArrivalTime         *string
	WorkingDepartureTime       *string
	WorkingPassingTime         *string
	PublicArrivalTime          *string
	PublicDepartureTime        *string
	RoutingDelay               int
	FalseDestinationLocationID *string

	CancellationReasonID             *int
	CancellationReasonLocationID     *string
	CancellationReasonIsNearLocation *bool
}

func (u UnitOfWork) scheduleToRecords(schedule unmarshaller.Schedule) (scheduleRecord, []scheduleLocationRecord, error) {
	var sRecord scheduleRecord
	var slRecords []scheduleLocationRecord

	sRecord.ID = uuid.New()
	sRecord.MessageID = u.messageID
	sRecord.TimetableID = u.timetableID
	sRecord.ScheduleID = schedule.RID
	sRecord.UID = schedule.UID
	sRecord.ScheduledStartDate = schedule.ScheduledStartDate
	sRecord.Headcode = schedule.Headcode
	sRecord.RetailServiceID = *schedule.RetailServiceID
	sRecord.TOCid = schedule.TOC
	sRecord.Service = string(schedule.Service)
	sRecord.Category = string(schedule.Category)
	sRecord.IsPassengerService = schedule.PassengerService
	sRecord.IsActive = schedule.Active
	sRecord.IsDeleted = schedule.Deleted
	sRecord.IsCharter = schedule.Charter
	if schedule.CancellationReason != nil {
		sRecord.IsCancelled = true // Added to maintain compatibility with timetable journeys.
		sRecord.CancellationReasonID = &schedule.CancellationReason.ReasonID
		sRecord.CancellationReasonLocationID = schedule.CancellationReason.TIPLOC
		sRecord.CancellationReasonIsNearLocation = &schedule.CancellationReason.Near
	}
	sRecord.DivertedViaLocationID = schedule.DivertedVia
	if schedule.DiversionReason != nil {
		sRecord.DiversionReasonID = &schedule.DiversionReason.ReasonID
		sRecord.DiversionReasonLocationID = schedule.DiversionReason.TIPLOC
		sRecord.DiversionReasonIsNearLocation = &schedule.DiversionReason.Near
	}

	var previousFormationID *string
	for _, location := range schedule.Locations {
		var slRecord scheduleLocationRecord
		slRecord.ID = uuid.New()
		slRecord.ScheduleUUID = sRecord.ID
		slRecord.ScheduleID = schedule.RID

		var base unmarshaller.LocationBase
		slRecord.Type = string(location.Type)
		switch location.Type {
		case unmarshaller.LocationTypeOrigin:
			loc := location.Origin
			base = loc.LocationBase
			slRecord.WorkingDepartureTime = &loc.WorkingDepartureTime
			slRecord.WorkingArrivalTime = loc.WorkingArrivalTime
			slRecord.PublicArrivalTime = loc.PublicArrivalTime
			slRecord.PublicDepartureTime = loc.PublicDepartureTime
			slRecord.FalseDestinationLocationID = loc.FalseDestination
		case unmarshaller.LocationTypeOperationalOrigin:
			loc := location.OperationalOrigin
			base = loc.LocationBase
			slRecord.WorkingDepartureTime = &loc.WorkingDepartureTime
			slRecord.WorkingArrivalTime = loc.WorkingArrivalTime
		case unmarshaller.LocationTypeIntermediate:
			loc := location.Intermediate
			base = loc.LocationBase
			slRecord.WorkingArrivalTime = &loc.WorkingArrivalTime
			slRecord.WorkingDepartureTime = &loc.WorkingDepartureTime
			slRecord.PublicArrivalTime = loc.PublicArrivalTime
			slRecord.PublicDepartureTime = loc.PublicDepartureTime
			slRecord.FalseDestinationLocationID = loc.FalseDestination
			slRecord.RoutingDelay = loc.RoutingDelay
		case unmarshaller.LocationTypeOperationalIntermediate:
			loc := location.OperationalIntermediate
			base = loc.LocationBase
			slRecord.WorkingArrivalTime = &loc.WorkingArrivalTime
			slRecord.WorkingDepartureTime = &loc.WorkingDepartureTime
			slRecord.RoutingDelay = loc.RoutingDelay
		case unmarshaller.LocationTypeIntermediatePassing:
			loc := location.IntermediatePassing
			base = loc.LocationBase
			slRecord.WorkingPassingTime = &loc.WorkingPassingTime
			slRecord.RoutingDelay = loc.RoutingDelay
		case unmarshaller.LocationTypeDestination:
			loc := location.Destination
			base = loc.LocationBase
			slRecord.WorkingArrivalTime = &loc.WorkingArrivalTime
			slRecord.WorkingDepartureTime = loc.WorkingDepartureTime
			slRecord.RoutingDelay = loc.RoutingDelay
		case unmarshaller.LocationTypeOperationalDestination:
			loc := location.OperationalDestination
			base = loc.LocationBase
			slRecord.WorkingArrivalTime = &loc.WorkingArrivalTime
			slRecord.WorkingDepartureTime = loc.WorkingDepartureTime
			slRecord.RoutingDelay = loc.RoutingDelay
		default:
			return sRecord, nil, fmt.Errorf("unknown location type %s", location.Type)
		}

		slRecord.LocationID = base.TIPLOC
		if base.Activities != nil {
			slRecord.Activities = sliceActivities(*base.Activities)
		}
		if base.PlannedActivities != nil {
			slRecord.PlannedActivities = sliceActivities(*base.PlannedActivities)
		}
		slRecord.IsCancelled = base.Cancelled
		slRecord.FormationID = base.FormationID
		slRecord.IsAffectedByDiversion = base.AffectedByDiversion

		if base.CancellationReason != nil {
			slRecord.CancellationReasonID = &base.CancellationReason.ReasonID
			slRecord.CancellationReasonLocationID = base.CancellationReason.TIPLOC
			slRecord.CancellationReasonIsNearLocation = &base.CancellationReason.Near
		}

		// Deal with FormationID 'rippling' rules.
		if slRecord.IsCancelled {
			previousFormationID = nil
		}
		if slRecord.FormationID == nil && previousFormationID != nil {
			slRecord.FormationID = previousFormationID
		}
		if slRecord.FormationID != nil {
			previousFormationID = slRecord.FormationID
		}
		slRecords = append(slRecords, slRecord)
	}
	return sRecord, slRecords, nil
}

// sliceActivities takes a string of 2-character activity codes and returns a slice of those codes (as strings).
// See railreader.ActivityCode for valid codes and their meanings.
func sliceActivities(activities string) []string {
	var slice []string
	if activities == "" {
		slice = append(slice, "  ")
		return slice
	}
	for i := 0; i < len(activities); i += 2 {
		slice = append(slice, activities[i:i+2])
	}
	return slice
}

func (u UnitOfWork) insertScheduleRecord(record scheduleRecord) error {
	_, err := u.tx.Exec(u.ctx, `
		INSERT INTO darwin.schedules (
			id,
			message_id,
			timetable_id,
			schedule_id,
			uid,
			scheduled_start_date,
			headcode,
			retail_service_id,
			toc_id,
			service,
			category,
			is_passenger_service,
			is_active,
			is_deleted,
			is_charter,
			cancellation_reason_id,
			cancellation_reason_location_id,
			cancellation_reason_is_near_location,
			diverted_via_location_id,
			diversion_reason_id,
			diversion_reason_location_id,
			diversion_reason_is_near_location,
			is_cancelled
		) VALUES (
			@id,
			@message_id,
			@timetable_id,
			@schedule_id,
			@uid,
			@scheduled_start_date,
			@headcode,
			@retail_service_id,
			@toc_id,
			@service,
			@category,
			@is_passenger_service,
			@is_active,
			@is_deleted,
			@is_charter,
			@cancellation_reason_id,
			@cancellation_reason_location_id,
			@cancellation_reason_is_near_location,
			@diverted_via_location_id,
			@diversion_reason_id,
			@diversion_reason_location_id,
			@diversion_reason_is_near_location,
			@is_cancelled
		);
		`, pgx.StrictNamedArgs{
		"id":                                   record.ID,
		"message_id":                           record.MessageID,
		"timetable_id":                         record.TimetableID,
		"schedule_id":                          record.ScheduleID,
		"uid":                                  record.UID,
		"scheduled_start_date":                 record.ScheduledStartDate,
		"headcode":                             record.Headcode,
		"retail_service_id":                    record.RetailServiceID,
		"toc_id":                               record.TOCid,
		"service":                              record.Service,
		"category":                             record.Category,
		"is_passenger_service":                 record.IsPassengerService,
		"is_active":                            record.IsActive,
		"is_deleted":                           record.IsDeleted,
		"is_charter":                           record.IsCharter,
		"cancellation_reason_id":               record.CancellationReasonID,
		"cancellation_reason_location_id":      record.CancellationReasonLocationID,
		"cancellation_reason_is_near_location": record.CancellationReasonIsNearLocation,
		"diverted_via_location_id":             record.DivertedViaLocationID,
		"diversion_reason_id":                  record.DiversionReasonID,
		"diversion_reason_location_id":         record.DiversionReasonLocationID,
		"diversion_reason_is_near_location":    record.DiversionReasonIsNearLocation,
		"is_cancelled":                         record.IsCancelled,
	})
	if err != nil {
		return err
	}
	return nil
}

func (u UnitOfWork) insertScheduleLocationRecords(records []scheduleLocationRecord) error {
	batch := &pgx.Batch{}
	for _, record := range records {
		batch.Queue(`
			INSERT INTO darwin.schedule_locations (
				id,
				schedule_uuid,
				schedule_id,
				location_id,
				activities,
				planned_activities,
				is_cancelled,
				formation_id,
				is_affected_by_diversion,
				type,
				working_arrival_time,
				working_departure_time,
				working_passing_time,
				public_arrival_time,
				public_departure_time,
				routing_delay,
				false_destination_location_id,
				cancellation_reason_id,
				cancellation_reason_location_id,
				cancellation_reason_is_near_location
			) VALUES (
				@id,
				@schedule_uuid,
				@schedule_id,
				@location_id,
				@activities,
				@planned_activities,
				@is_cancelled,
				@formation_id,
				@is_affected_by_diversion,
				@type,
				@working_arrival_time,
				@working_departure_time,
				@working_passing_time,
				@public_arrival_time,
				@public_departure_time,
				@routing_delay,
				@false_destination_location_id,
				@cancellation_reason_id,
				@cancellation_reason_location_id,
				@cancellation_reason_is_near_location
			);
		`, pgx.StrictNamedArgs{
			"id":                                   record.ID,
			"schedule_uuid":                        record.ScheduleUUID,
			"schedule_id":                          record.ScheduleID,
			"location_id":                          record.LocationID,
			"activities":                           record.Activities,
			"planned_activities":                   record.PlannedActivities,
			"is_cancelled":                         record.IsCancelled,
			"formation_id":                         record.FormationID,
			"is_affected_by_diversion":             record.IsAffectedByDiversion,
			"type":                                 record.Type,
			"working_arrival_time":                 record.WorkingArrivalTime,
			"working_departure_time":               record.WorkingDepartureTime,
			"working_passing_time":                 record.WorkingPassingTime,
			"public_arrival_time":                  record.PublicArrivalTime,
			"public_departure_time":                record.PublicDepartureTime,
			"routing_delay":                        record.RoutingDelay,
			"false_destination_location_id":        record.FalseDestinationLocationID,
			"cancellation_reason_id":               record.CancellationReasonID,
			"cancellation_reason_location_id":      record.CancellationReasonLocationID,
			"cancellation_reason_is_near_location": record.CancellationReasonIsNearLocation,
		})
	}
	results := u.tx.SendBatch(u.ctx, batch)
	return results.Close()
}
