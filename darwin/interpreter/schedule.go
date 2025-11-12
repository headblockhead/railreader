package interpreter

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/headblockhead/railreader/darwin/unmarshaller"
)

func (u UnitOfWork) interpretSchedule(schedule unmarshaller.Schedule) error {
	scheduleRecord, scheduleLocationRecords, err := u.scheduleToRecords(schedule)
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

// TODO: add insertion functions
