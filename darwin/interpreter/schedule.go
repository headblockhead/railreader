package interpreter

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/headblockhead/railreader/darwin/repository"
	"github.com/headblockhead/railreader/darwin/unmarshaller"
)

func (u UnitOfWork) interpretSchedule(schedule unmarshaller.Schedule) error {
	var row scheduleRecord
	row.ScheduleID = schedule.RID
	row.MessageID = u.messageID
	row.UID = schedule.UID
	location, err := time.LoadLocation("Europe/London")
	if err != nil {
		return fmt.Errorf("failed to load time location Europe/London: %w", err)
	}
	startDate, err := time.ParseInLocation("2006-01-02", schedule.ScheduledStartDate, location)
	if err != nil {
		return fmt.Errorf("failed to parse ScheduledStartDate %q: %w", schedule.ScheduledStartDate, err)
	}
	row.ScheduledStartDate = startDate

	row.Headcode = schedule.Headcode
	row.RetailServiceID = *schedule.RetailServiceID
	row.TOCid = schedule.TOC
	row.Service = string(schedule.Service)
	row.Category = string(schedule.Category)
	row.IsPassengerService = schedule.PassengerService
	row.IsActive = schedule.Active
	row.IsDeleted = schedule.Deleted
	row.IsCharter = schedule.Charter
	if schedule.CancellationReason != nil {
		row.IsCancelled = true // Added to maintain compatibility with timetable journeys.
		row.CancellationReasonID = &schedule.CancellationReason.ReasonID
		row.CancellationReasonLocationID = schedule.CancellationReason.TIPLOC
		row.CancellationReasonIsNearLocation = &schedule.CancellationReason.Near
	}
	row.DivertedViaLocationID = schedule.DivertedVia
	if schedule.DiversionReason != nil {
		row.DiversionReasonID = &schedule.DiversionReason.ReasonID
		row.DiversionReasonLocationID = schedule.DiversionReason.TIPLOC
		row.DiversionReasonIsNearLocation = &schedule.DiversionReason.Near
	}

	locationRows, err := scheduleLocationsToRows(schedule.Locations, schedule.RID)
	if err != nil {
		return fmt.Errorf("failed to convert schedule locations to rows: %w", err)
	}

	return nil
}

type scheduleRecord struct {
	ID uuid.UUID

	MessageID   *string
	TimetableID *string

	ScheduleID         string
	UID                string
	ScheduledStartDate time.Time

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

	// TODO

	for _, sch := range schedule.Locations {
		
	}
}

func scheduleLocationsToRows(locations []unmarshaller.ScheduleLocation, scheduleID string) ([]scheduleLocationRecord, error) {
	var records []scheduleLocationRecord
	var previousFormationID *string

	for _, location := range locations {
		var record scheduleLocationRecord
		switch location.Type {
		case unmarshaller.LocationTypeOrigin:
			loc := location.Origin
			record = baseScheduleToRecord(loc.LocationBase)
			record.WorkingDepartureTime = &loc.WorkingDepartureTime
			record.WorkingArrivalTime = loc.WorkingArrivalTime
			record.PublicArrivalTime = loc.PublicArrivalTime
			record.PublicDepartureTime = loc.PublicDepartureTime
			record.FalseDestinationLocationID = loc.FalseDestination
		case unmarshaller.LocationTypeOperationalOrigin:
			loc := location.OperationalOrigin
			record = baseScheduleToRecord(loc.LocationBase)
			record.WorkingDepartureTime = &loc.WorkingDepartureTime
			record.WorkingArrivalTime = loc.WorkingArrivalTime
		case unmarshaller.LocationTypeIntermediate:
			loc := location.Intermediate
			record = baseScheduleToRecord(loc.LocationBase)
			record.WorkingArrivalTime = &loc.WorkingArrivalTime
			record.WorkingDepartureTime = &loc.WorkingDepartureTime
			record.PublicArrivalTime = loc.PublicArrivalTime
			record.PublicDepartureTime = loc.PublicDepartureTime
			record.FalseDestinationLocationID = loc.FalseDestination
			record.RoutingDelay = loc.RoutingDelay
		case unmarshaller.LocationTypeOperationalIntermediate:
			loc := location.OperationalIntermediate
			record = baseScheduleToRecord(loc.LocationBase)
			record.WorkingArrivalTime = &loc.WorkingArrivalTime
			record.WorkingDepartureTime = &loc.WorkingDepartureTime
			record.RoutingDelay = loc.RoutingDelay
		case unmarshaller.LocationTypeIntermediatePassing:
			loc := location.IntermediatePassing
			record = baseScheduleToRecord(loc.LocationBase)
			record.WorkingPassingTime = &loc.WorkingPassingTime
			record.RoutingDelay = loc.RoutingDelay
		case unmarshaller.LocationTypeDestination:
			loc := location.Destination
			record = baseScheduleToRecord(loc.LocationBase)
			record.WorkingArrivalTime = &loc.WorkingArrivalTime
			record.WorkingDepartureTime = loc.WorkingDepartureTime
			record.RoutingDelay = loc.RoutingDelay
		case unmarshaller.LocationTypeOperationalDestination:
			loc := location.OperationalDestination
			record = baseScheduleToRecord(loc.LocationBase)
			record.WorkingArrivalTime = &loc.WorkingArrivalTime
			record.WorkingDepartureTime = loc.WorkingDepartureTime
			record.RoutingDelay = loc.RoutingDelay
		default:
			return nil, fmt.Errorf("unknown location type %s", location.Type)
		}

		record.ID = uuid.New()
		record.ScheduleID = scheduleID
		record.ScheduleUUID = 

		record.Type = string(location.Type)

		// Deal with FormationID 'rippling' rules.
		if record.IsCancelled {
			previousFormationID = nil
		}
		if record.FormationID == nil && previousFormationID != nil {
			record.FormationID = previousFormationID
		}
		if record.FormationID != nil {
			previousFormationID = record.FormationID
		}
		records = append(records, record)
	}
	return records, nil
}

func baseScheduleToRecord(base unmarshaller.LocationBase) scheduleLocationRecord {
	var record repository.ScheduleLocationRow
	record.LocationID = base.TIPLOC
	if base.Activities != nil {
		record.Activities = sliceActivities(*base.Activities)
	}
	if base.PlannedActivities != nil {
		record.PlannedActivities = sliceActivities(*base.PlannedActivities)
	}
	record.IsCancelled = base.Cancelled
	record.FormationID = base.FormationID
	record.IsAffectedByDiversion = base.AffectedByDiversion
	if base.CancellationReason != nil {
		record.CancellationReasonID = &base.CancellationReason.ReasonID
		record.CancellationReasonLocationID = base.CancellationReason.TIPLOC
		record.CancellationReasonIsNearLocation = &base.CancellationReason.Near
	}
	return record
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
