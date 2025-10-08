package interpreter

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/headblockhead/railreader/darwin/repository"
	"github.com/headblockhead/railreader/darwin/unmarshaller"
)

func (u UnitOfWork) GetLastTimetable() (*repository.TimetableRow, error) {
	return u.timetableRepository.SelectLast()
}

func (u *UnitOfWork) InterpretTimetable(timetable unmarshaller.Timetable, filename string) error {
	u.timetableID = &timetable.ID
	log := u.log.With(slog.String("timetable_id", timetable.ID))
	log.Debug("interpreting a Timetable")
	location, err := time.LoadLocation("Europe/London")
	if err != nil {
		return fmt.Errorf("failed to load time location: %w", err)
	}

	if err := u.timetableRepository.Insert(repository.TimetableRow{
		TimetableID:     timetable.ID,
		FirstReceivedAt: time.Now().In(location),
		Filename:        filename,
	}); err != nil {
		return fmt.Errorf("failed to insert timetable into repository: %w", err)
	}

	var scheduleRows []repository.ScheduleRow
	var scheduleLocationRows []repository.ScheduleLocationRow
	for _, journey := range timetable.Journeys {
		scheduleRow, locationRows, err := journeyAndLocationsToRows(journey, timetable.ID)
		if err != nil {
			return fmt.Errorf("failed to convert journey and locations to rows: %w", err)
		}
		scheduleRows = append(scheduleRows, scheduleRow)
		scheduleLocationRows = append(scheduleLocationRows, locationRows...)
	}

	if err := u.scheduleRepository.InsertMany(scheduleRows); err != nil {
		return fmt.Errorf("failed to insert schedule rows into repository: %w", err)
	}
	if err := u.scheduleLocationRepository.InsertMany(scheduleLocationRows); err != nil {
		return fmt.Errorf("failed to insert schedule location rows into repository: %w", err)
	}

	var associationRows []repository.AssociationRow
	for _, association := range timetable.Associations {
		row, err := u.associationToRow(association)
		if err != nil {
			return fmt.Errorf("failed to convert association to row: %w", err)
		}
		associationRows = append(associationRows, row)
	}

	if err := u.associationRepository.InsertMany(associationRows); err != nil {
		return fmt.Errorf("failed to insert schedule association rows into repository: %w", err)
	}

	return nil
}

func journeyAndLocationsToRows(journey unmarshaller.Journey, timetableID string) (repository.ScheduleRow, []repository.ScheduleLocationRow, error) {
	row, err := journeyToScheduleRow(journey, timetableID)
	if err != nil {
		return row, nil, fmt.Errorf("failed to convert journey to schedule row: %w", err)
	}
	locationRows, err := journeyLocationsToRows(journey.Locations, journey.RID)
	if err != nil {
		return row, nil, fmt.Errorf("failed to convert journey locations to rows: %w", err)
	}
	return row, locationRows, nil
}

func journeyToScheduleRow(journey unmarshaller.Journey, timetableID string) (repository.ScheduleRow, error) {
	var row repository.ScheduleRow
	row.ScheduleID = journey.RID
	row.TimetableID = &timetableID
	row.UID = journey.UID
	location, err := time.LoadLocation("Europe/London")
	if err != nil {
		return row, fmt.Errorf("failed to load time location Europe/London: %w", err)
	}
	startDate, err := time.ParseInLocation("2006-01-02", journey.ScheduledStartDate, location)
	if err != nil {
		return row, fmt.Errorf("failed to parse ScheduledStartDate %q: %w", journey.ScheduledStartDate, err)
	}
	row.ScheduledStartDate = startDate

	row.Headcode = journey.Headcode
	row.TrainOperatingCompanyID = journey.TOC
	row.Service = string(journey.Service)
	row.Category = string(journey.Category)
	row.IsPassengerService = journey.PassengerService
	row.IsActive = !journey.QTrain
	row.IsDeleted = journey.Deleted
	row.IsCharter = journey.Charter
	row.IsCancelled = journey.Cancelled
	if journey.CancellationReason != nil {
		row.CancellationReasonID = &journey.CancellationReason.ReasonID
		row.CancellationReasonLocationID = journey.CancellationReason.TIPLOC
		row.CancellationReasonIsNearLocation = &journey.CancellationReason.Near
	}
	return row, nil
}

func journeyLocationsToRows(locations []unmarshaller.JourneyLocation, scheduleID string) ([]repository.ScheduleLocationRow, error) {
	var rows []repository.ScheduleLocationRow
	for i, location := range locations {
		var row repository.ScheduleLocationRow
		switch location.Type {
		case unmarshaller.LocationTypeOrigin:
			loc := location.Origin
			row = newLocationRowFromJourneyBase(loc.JourneyLocationBase, i)
			row.WorkingDepartureTime = &loc.WorkingDepartureTime
			row.WorkingArrivalTime = loc.WorkingArrivalTime
			row.PublicArrivalTime = loc.PublicArrivalTime
			row.PublicDepartureTime = loc.PublicDepartureTime
			row.FalseDestinationLocationID = loc.FalseDestination
		case unmarshaller.LocationTypeOperationalOrigin:
			loc := location.OperationalOrigin
			row = newLocationRowFromJourneyBase(loc.JourneyLocationBase, i)
			row.WorkingDepartureTime = &loc.WorkingDepartureTime
			row.WorkingArrivalTime = loc.WorkingArrivalTime
		case unmarshaller.LocationTypeIntermediate:
			loc := location.Intermediate
			row = newLocationRowFromJourneyBase(loc.JourneyLocationBase, i)
			row.WorkingArrivalTime = &loc.WorkingArrivalTime
			row.WorkingDepartureTime = &loc.WorkingDepartureTime
			row.PublicArrivalTime = loc.PublicArrivalTime
			row.PublicDepartureTime = loc.PublicDepartureTime
			row.FalseDestinationLocationID = loc.FalseDestination
			row.RoutingDelay = loc.RoutingDelay
		case unmarshaller.LocationTypeOperationalIntermediate:
			loc := location.OperationalIntermediate
			row = newLocationRowFromJourneyBase(loc.JourneyLocationBase, i)
			row.WorkingArrivalTime = &loc.WorkingArrivalTime
			row.WorkingDepartureTime = &loc.WorkingDepartureTime
			row.RoutingDelay = loc.RoutingDelay
		case unmarshaller.LocationTypeIntermediatePassing:
			loc := location.IntermediatePassing
			row = newLocationRowFromJourneyBase(loc.JourneyLocationBase, i)
			row.WorkingPassingTime = &loc.WorkingPassingTime
			row.RoutingDelay = loc.RoutingDelay
		case unmarshaller.LocationTypeDestination:
			loc := location.Destination
			row = newLocationRowFromJourneyBase(loc.JourneyLocationBase, i)
			row.WorkingArrivalTime = &loc.WorkingArrivalTime
			row.WorkingDepartureTime = loc.WorkingDepartureTime
			row.RoutingDelay = loc.RoutingDelay
		case unmarshaller.LocationTypeOperationalDestination:
			loc := location.OperationalDestination
			row = newLocationRowFromJourneyBase(loc.JourneyLocationBase, i)
			row.WorkingArrivalTime = &loc.WorkingArrivalTime
			row.WorkingDepartureTime = loc.WorkingDepartureTime
			row.RoutingDelay = loc.RoutingDelay
		default:
			return nil, fmt.Errorf("unknown location type %s", location.Type)
		}
		row.Type = string(location.Type)
		row.ScheduleID = scheduleID
		rows = append(rows, row)
	}
	return rows, nil
}

func newLocationRowFromJourneyBase(baseValues unmarshaller.JourneyLocationBase, sequence int) repository.ScheduleLocationRow {
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
	row.Platform = baseValues.Platform
	return row
}
