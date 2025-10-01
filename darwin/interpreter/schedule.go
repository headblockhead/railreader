package interpreter

import (
	"fmt"
	"time"

	"github.com/headblockhead/railreader/darwin/repository"
	"github.com/headblockhead/railreader/darwin/unmarshaller"
)

func (u UnitOfWork) InterpretSchedule(schedule unmarshaller.Schedule) error {
	// Delete existing schedule with same RID, if any.
	if err := u.scheduleRepository.Delete(schedule.RID); err != nil {
		return fmt.Errorf("failed to delete existing schedule with RID %q: %w", schedule.RID, err)
	}

	var row repository.ScheduleRow
	row.ScheduleID = schedule.RID
	row.MessageID = &u.messageID
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
	row.RetailServiceID = schedule.RetailServiceID
	row.TrainOperatingCompanyID = schedule.TOC
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
		row.LateReasonID = &schedule.DiversionReason.ReasonID
		row.LateReasonLocationID = schedule.DiversionReason.TIPLOC
		row.LateReasonIsNearLocation = &schedule.DiversionReason.Near
	}

	locationRows, err := scheduleLocationsToRows(schedule.Locations, schedule.RID)
	if err != nil {
		return fmt.Errorf("failed to convert schedule locations to rows: %w", err)
	}

	if err := u.scheduleRepository.Insert(row); err != nil {
		return fmt.Errorf("failed to insert schedule into repository: %w", err)
	}
	if err := u.scheduleLocationRepository.InsertMany(locationRows); err != nil {
		return fmt.Errorf("failed to insert schedule locations into repository: %w", err)
	}
	return nil
}

func scheduleLocationsToRows(locations []unmarshaller.ScheduleLocation, scheduleID string) ([]repository.ScheduleLocationRow, error) {
	var rows []repository.ScheduleLocationRow
	var previousFormationID *string

	for i, location := range locations {
		var row repository.ScheduleLocationRow
		switch location.Type {
		case unmarshaller.LocationTypeOrigin:
			loc := location.Origin
			row = newLocationRowFromBase(loc.LocationBase, i)
			row.WorkingDepartureTime = &loc.WorkingDepartureTime
			row.WorkingArrivalTime = loc.WorkingArrivalTime
			row.PublicArrivalTime = loc.PublicArrivalTime
			row.PublicDepartureTime = loc.PublicDepartureTime
			row.FalseDestinationLocationID = loc.FalseDestination
		case unmarshaller.LocationTypeOperationalOrigin:
			loc := location.OperationalOrigin
			row = newLocationRowFromBase(loc.LocationBase, i)
			row.WorkingDepartureTime = &loc.WorkingDepartureTime
			row.WorkingArrivalTime = loc.WorkingArrivalTime
		case unmarshaller.LocationTypeIntermediate:
			loc := location.Intermediate
			row = newLocationRowFromBase(loc.LocationBase, i)
			row.WorkingArrivalTime = &loc.WorkingArrivalTime
			row.WorkingDepartureTime = &loc.WorkingDepartureTime
			row.PublicArrivalTime = loc.PublicArrivalTime
			row.PublicDepartureTime = loc.PublicDepartureTime
			row.FalseDestinationLocationID = loc.FalseDestination
			row.RoutingDelay = loc.RoutingDelay
		case unmarshaller.LocationTypeOperationalIntermediate:
			loc := location.OperationalIntermediate
			row = newLocationRowFromBase(loc.LocationBase, i)
			row.WorkingArrivalTime = &loc.WorkingArrivalTime
			row.WorkingDepartureTime = &loc.WorkingDepartureTime
			row.RoutingDelay = loc.RoutingDelay
		case unmarshaller.LocationTypeIntermediatePassing:
			loc := location.IntermediatePassing
			row = newLocationRowFromBase(loc.LocationBase, i)
			row.WorkingPassingTime = &loc.WorkingPassingTime
			row.RoutingDelay = loc.RoutingDelay
		case unmarshaller.LocationTypeDestination:
			loc := location.Destination
			row = newLocationRowFromBase(loc.LocationBase, i)
			row.WorkingArrivalTime = &loc.WorkingArrivalTime
			row.WorkingDepartureTime = loc.WorkingDepartureTime
			row.RoutingDelay = loc.RoutingDelay
		case unmarshaller.LocationTypeOperationalDestination:
			loc := location.OperationalDestination
			row = newLocationRowFromBase(loc.LocationBase, i)
			row.WorkingArrivalTime = &loc.WorkingArrivalTime
			row.WorkingDepartureTime = loc.WorkingDepartureTime
			row.RoutingDelay = loc.RoutingDelay
		default:
			return nil, fmt.Errorf("unknown location type %s", location.Type)
		}
		row.Type = string(location.Type)
		row.ScheduleID = scheduleID

		// Deal with FormationID 'rippling' rules.
		if row.IsCancelled {
			previousFormationID = nil
		}
		if row.FormationID == nil && previousFormationID != nil {
			row.FormationID = previousFormationID
		}
		if row.FormationID != nil {
			previousFormationID = row.FormationID
		}
		rows = append(rows, row)
	}
	return rows, nil
}

func newLocationRowFromBase(baseValues unmarshaller.LocationBase, sequence int) repository.ScheduleLocationRow {
	var row repository.ScheduleLocationRow
	row.Sequence = sequence
	row.LocationID = baseValues.TIPLOC
	if baseValues.Activities != nil {
		row.Activities = sliceActivities(*baseValues.Activities)
	}
	if baseValues.PlannedActivities != nil {
		row.PlannedActivities = sliceActivities(*baseValues.PlannedActivities)
	}
	row.IsCancelled = baseValues.Cancelled
	row.FormationID = baseValues.FormationID
	row.IsAffectedByDiversion = baseValues.AffectedByDiversion
	if baseValues.CancellationReason != nil {
		row.CancellationReasonID = &baseValues.CancellationReason.ReasonID
		row.CancellationReasonLocationID = baseValues.CancellationReason.TIPLOC
		row.CancellationReasonIsNearLocation = &baseValues.CancellationReason.Near
	}
	return row
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
