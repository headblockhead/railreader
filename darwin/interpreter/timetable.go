package interpreter

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/headblockhead/railreader/darwin/repository"
	"github.com/headblockhead/railreader/darwin/unmarshaller"
)

// TODO: tidy this generally, and figure out how to use copy to make this not take several minutes.

func (u UnitOfWork) InterpretTimetable(timetable unmarshaller.Timetable) error {
	log := u.log.With(slog.String("timetable_id", timetable.ID))
	log.Debug("interpreting a Timetable")
	location, err := time.LoadLocation("Europe/London")
	if err != nil {
		return fmt.Errorf("failed to load time location: %w", err)
	}

	if err := u.timetableRepository.Insert(repository.TimetableRow{
		TimetableID:     timetable.ID,
		FirstReceivedAt: time.Now().In(location),
	}); err != nil {
		return fmt.Errorf("failed to insert timetable into repository: %w", err)
	}

	for _, journey := range timetable.Journeys {
		if err := u.interpretJourney(timetable.ID, journey); err != nil {
			return fmt.Errorf("failed to interpret journey %q: %w", journey.RID, err)
		}
	}

	for _, association := range timetable.Associations {
	if err := u.associationRepository.Insert(repository.AssociationRow{
	}); err != nil {
		return fmt.Errorf("failed to insert timetable into repository: %w", err)
	}
}

	return nil
}

func (u UnitOfWork) interpretJourney(timetableID string, journey unmarshaller.Journey) error {
	log := u.log.With(slog.String("schedule_id", journey.RID))
	log.Debug("interpreting a Journey")
	var row repository.ScheduleRow
	row.ScheduleID = journey.RID
	row.TimetableID = &timetableID
	row.UID = journey.UID
	log.Debug("parsing ScheduledStartDate", slog.String("value", journey.ScheduledStartDate))
	location, err := time.LoadLocation("Europe/London")
	if err != nil {
		return fmt.Errorf("failed to load time location Europe/London: %w", err)
	}
	startDate, err := time.ParseInLocation("2006-01-02", journey.ScheduledStartDate, location)
	if err != nil {
		return fmt.Errorf("failed to parse ScheduledStartDate %q: %w", journey.ScheduledStartDate, err)
	}
	log.Debug("parsed ScheduledStartDate", slog.Time("time", startDate))
	row.ScheduledStartDate = startDate
	row.Headcode = journey.Headcode
	row.TrainOperatingCompanyID = string(journey.TOC)
	row.Service = string(journey.Service)
	row.Category = string(journey.Category)
	row.PassengerService = journey.PassengerService
	// TODO: is active supposed to be true for timetable entries?
	row.Active = true
	row.Deleted = journey.Deleted
	row.Charter = journey.Charter
	if journey.CancellationReason != nil {
		log.Debug("CancellationReason is set")
		row.Cancelled = true
		row.CancellationReasonID = &journey.CancellationReason.ReasonID
		if journey.CancellationReason.TIPLOC != nil && *journey.CancellationReason.TIPLOC != "" {
			log.Debug("CancellationReason.TIPLOC is set")
			tiploc := string(*journey.CancellationReason.TIPLOC)
			row.CancellationReasonLocationID = &tiploc
		}
		row.CancellationReasonNearLocation = &journey.CancellationReason.Near
	}

	previousTime := time.Time{}
	var locationRows []repository.ScheduleLocationRow
	for sequence, timetableLocation := range journey.Locations {
		locationLog := log.With(slog.Int("sequence", sequence), slog.String("type", string(timetableLocation.Type)))
		locationLog.Debug("parsing journey location")
		locationRow, nextTime, err := convertTimetableLocationToRow(locationLog, journey.RID, sequence, startDate, previousTime, timetableLocation)
		if err != nil {
			return fmt.Errorf("failed to parse schedule location at sequence %d: %w", sequence, err)
		}
		previousTime = nextTime
		locationRows = append(locationRows, locationRow)
	}

	if err := u.scheduleRepository.Insert(row); err != nil {
		return fmt.Errorf("failed to insert schedule into repository: %w", err)
	}
	if err := u.scheduleLocationRepository.InsertMany(locationRows); err != nil {
		return fmt.Errorf("failed to insert schedule locations into repository: %w", err)
	}
	return nil
}

type databaseableJourneyLocation interface {
	convertToDatabaseLocation(sequence int, previousTime time.Time, startDate time.Time) (databaseLocation repository.ScheduleLocationRow, nextTime time.Time, err error)
}

func convertTimetableLocationToRow(log *slog.Logger, scheduleID string, sequence int, startDate time.Time, previousTime time.Time, timetableLocation unmarshaller.TimetableLocation) (row repository.ScheduleLocationRow, nextTime time.Time, err error) {
	log.Debug("converting location")
	var location databaseableScheduleLocation
	switch timetableLocation.Type {
	case unmarshaller.LocationTypeOrigin:
		location = timetableOriginLocation{log: log, src: timetableLocation.Origin}
	case unmarshaller.LocationTypeOperationalOrigin:
		location = timetableOperationalOriginLocation{log: log, src: timetableLocation.OperationalOrigin}
	case unmarshaller.LocationTypeIntermediate:
		location = timetableIntermediateLocation{log: log, src: timetableLocation.Intermediate}
	case unmarshaller.LocationTypeOperationalIntermediate:
		location = timetableOperationalIntermediateLocation{log: log, src: timetableLocation.OperationalIntermediate}
	case unmarshaller.LocationTypeIntermediatePassing:
		location = timetableIntermediatePassingLocation{log: log, src: timetableLocation.IntermediatePassing}
	case unmarshaller.LocationTypeDestination:
		location = timetableDestinationLocation{log: log, src: timetableLocation.Destination}
	case unmarshaller.LocationTypeOperationalDestination:
		location = timetableOperationalDestinationLocation{log: log, src: timetableLocation.OperationalDestination}
	default:
		err = fmt.Errorf("unknown location type %s", timetableLocation.Type)
		return
	}
	row, nextTime, err = location.convertToDatabaseLocation(sequence, previousTime, startDate)
	if err != nil {
		err = fmt.Errorf("failed to convert location to database location: %w", err)
		return
	}
	row.ScheduleID = scheduleID
	if row.Cancelled {
		log.Debug("location is marked as cancelled")
		return
	}
	return
}

func newDatabaseLocationWithBaseTimetableValues(log *slog.Logger, baseValues unmarshaller.TimetableLocationBase, sequence int) (row repository.ScheduleLocationRow) {
	row.Sequence = sequence
	row.LocationID = string(baseValues.TIPLOC)
	if baseValues.Activities != nil {
		log.Debug("Activities is set")
		activities := activitiesToSlice(*baseValues.Activities)
		row.Activities = &activities
	}
	if baseValues.PlannedActivities != nil {
		log.Debug("PlannedActivities is set")
		plannedActivities := activitiesToSlice(*baseValues.PlannedActivities)
		row.PlannedActivities = &plannedActivities
	}
	row.Cancelled = baseValues.Cancelled
	row.Platform = baseValues.Platform
	return
}

type timetableOriginLocation struct {
	log *slog.Logger
	// MUST contain: WTD
	// MAY contain: WTA, PTA, PTD, FD
	src *unmarshaller.OriginTimetableLocation
}

func (c timetableOriginLocation) convertToDatabaseLocation(sequence int, previousTime time.Time, startDate time.Time) (databaseLocation repository.ScheduleLocationRow, nextTime time.Time, err error) {
	databaseLocation = newDatabaseLocationWithBaseTimetableValues(c.log, c.src.TimetableLocationBase, sequence)
	databaseLocation.Type = string(unmarshaller.LocationTypeOrigin)
	databaseLocation.WorkingArrivalTime, previousTime, err = convertOptionalTrainTime(previousTime, startDate, c.src.WorkingArrivalTime)
	if err != nil {
		err = fmt.Errorf("failed to parse WorkingArrivalTime: %w", err)
		return
	}
	databaseLocation.WorkingDepartureTime, previousTime, err = convertTrainTime(previousTime, startDate, c.src.WorkingDepartureTime)
	if err != nil {
		err = fmt.Errorf("failed to parse WorkingDepartureTime: %w", err)
		return
	}
	databaseLocation.PublicArrivalTime, previousTime, err = convertOptionalTrainTime(previousTime, startDate, c.src.PublicArrivalTime)
	if err != nil {
		err = fmt.Errorf("failed to parse PublicArrivalTime: %w", err)
		return
	}
	databaseLocation.PublicDepartureTime, previousTime, err = convertOptionalTrainTime(previousTime, startDate, c.src.PublicDepartureTime)
	if err != nil {
		err = fmt.Errorf("failed to parse PublicDepartureTime: %w", err)
		return
	}
	databaseLocation.FalseDestinationLocationID = parseOptionalFalseDestination(c.src.FalseDestination)
	return databaseLocation, previousTime, nil
}

type timetableOperationalOriginLocation struct {
	log *slog.Logger
	// MUST contain: WTD
	// MAY contain: WTA
	src *unmarshaller.OperationalOriginTimetableLocation
}

func (c timetableOperationalOriginLocation) convertToDatabaseLocation(sequence int, previousTime time.Time, startDate time.Time) (databaseLocation repository.ScheduleLocationRow, nextTime time.Time, err error) {
	databaseLocation = newDatabaseLocationWithBaseTimetableValues(c.log, c.src.TimetableLocationBase, sequence)
	databaseLocation.Type = string(unmarshaller.LocationTypeOperationalOrigin)
	databaseLocation.WorkingArrivalTime, previousTime, err = convertOptionalTrainTime(previousTime, startDate, c.src.WorkingArrivalTime)
	if err != nil {
		err = fmt.Errorf("failed to parse WorkingArrivalTime: %w", err)
		return
	}
	databaseLocation.WorkingDepartureTime, previousTime, err = convertTrainTime(previousTime, startDate, c.src.WorkingDepartureTime)
	if err != nil {
		err = fmt.Errorf("failed to parse WorkingDepartureTime: %w", err)
		return
	}
	return databaseLocation, previousTime, nil
}

type timetableIntermediateLocation struct {
	log *slog.Logger
	// MUST contain: WTA, WTD
	// MAY contain: PTA, PTD, FD, RD
	src *unmarshaller.IntermediateTimetableLocation
}

func (c timetableIntermediateLocation) convertToDatabaseLocation(sequence int, previousTime time.Time, startDate time.Time) (databaseLocation repository.ScheduleLocationRow, nextTime time.Time, err error) {
	databaseLocation = newDatabaseLocationWithBaseTimetableValues(c.log, c.src.TimetableLocationBase, sequence)
	databaseLocation.Type = string(unmarshaller.LocationTypeIntermediate)
	databaseLocation.WorkingArrivalTime, previousTime, err = convertTrainTime(previousTime, startDate, c.src.WorkingArrivalTime)
	if err != nil {
		err = fmt.Errorf("failed to parse WorkingArrivalTime: %w", err)
		return
	}
	databaseLocation.WorkingDepartureTime, previousTime, err = convertTrainTime(previousTime, startDate, c.src.WorkingDepartureTime)
	if err != nil {
		err = fmt.Errorf("failed to parse WorkingDepartureTime: %w", err)
		return
	}
	databaseLocation.PublicArrivalTime, previousTime, err = convertOptionalTrainTime(previousTime, startDate, c.src.PublicArrivalTime)
	if err != nil {
		err = fmt.Errorf("failed to parse PublicArrivalTime: %w", err)
		return
	}
	databaseLocation.PublicDepartureTime, previousTime, err = convertOptionalTrainTime(previousTime, startDate, c.src.PublicDepartureTime)
	if err != nil {
		err = fmt.Errorf("failed to parse PublicDepartureTime: %w", err)
		return
	}
	databaseLocation.FalseDestinationLocationID = parseOptionalFalseDestination(c.src.FalseDestination)
	databaseLocation.RoutingDelay = parseOptionalRoutingDelay(c.src.RoutingDelay)
	return databaseLocation, previousTime, nil
}

type timetableOperationalIntermediateLocation struct {
	log *slog.Logger
	// MUST contain: WTA, WTD
	// MAY contain: RD
	src *unmarshaller.OperationalIntermediateTimetableLocation
}

func (c timetableOperationalIntermediateLocation) convertToDatabaseLocation(sequence int, previousTime time.Time, startDate time.Time) (databaseLocation repository.ScheduleLocationRow, nextTime time.Time, err error) {
	databaseLocation = newDatabaseLocationWithBaseTimetableValues(c.log, c.src.TimetableLocationBase, sequence)
	databaseLocation.Type = string(unmarshaller.LocationTypeOperationalIntermediate)
	databaseLocation.WorkingArrivalTime, previousTime, err = convertTrainTime(previousTime, startDate, c.src.WorkingArrivalTime)
	if err != nil {
		err = fmt.Errorf("failed to parse WorkingArrivalTime: %w", err)
		return
	}
	databaseLocation.WorkingDepartureTime, previousTime, err = convertTrainTime(previousTime, startDate, c.src.WorkingDepartureTime)
	if err != nil {
		err = fmt.Errorf("failed to parse WorkingDepartureTime: %w", err)
		return
	}
	databaseLocation.RoutingDelay = parseOptionalRoutingDelay(c.src.RoutingDelay)
	return databaseLocation, previousTime, nil
}

type timetableIntermediatePassingLocation struct {
	log *slog.Logger
	// MUST contain: WTP
	// MAY contain: RD
	src *unmarshaller.IntermediatePassingTimetableLocation
}

func (c timetableIntermediatePassingLocation) convertToDatabaseLocation(sequence int, previousTime time.Time, startDate time.Time) (databaseLocation repository.ScheduleLocationRow, nextTime time.Time, err error) {
	databaseLocation = newDatabaseLocationWithBaseTimetableValues(c.log, c.src.TimetableLocationBase, sequence)
	databaseLocation.Type = string(unmarshaller.LocationTypeIntermediatePassing)
	databaseLocation.WorkingPassingTime, previousTime, err = convertTrainTime(previousTime, startDate, c.src.WorkingPassingTime)
	if err != nil {
		err = fmt.Errorf("failed to parse WorkingPassingTime: %w", err)
		return
	}
	databaseLocation.RoutingDelay = parseOptionalRoutingDelay(c.src.RoutingDelay)
	return databaseLocation, previousTime, nil
}

type timetableDestinationLocation struct {
	log *slog.Logger
	// MUST contain: WTA
	// MAY contain: WTD, PTA, PTD, RD
	src *unmarshaller.DestinationTimetableLocation
}

func (c timetableDestinationLocation) convertToDatabaseLocation(sequence int, previousTime time.Time, startDate time.Time) (databaseLocation repository.ScheduleLocationRow, nextTime time.Time, err error) {
	databaseLocation = newDatabaseLocationWithBaseTimetableValues(c.log, c.src.TimetableLocationBase, sequence)
	databaseLocation.Type = string(unmarshaller.LocationTypeDestination)
	databaseLocation.WorkingArrivalTime, previousTime, err = convertTrainTime(previousTime, startDate, c.src.WorkingArrivalTime)
	if err != nil {
		err = fmt.Errorf("failed to parse WorkingArrivalTime: %w", err)
		return
	}
	databaseLocation.WorkingDepartureTime, previousTime, err = convertOptionalTrainTime(previousTime, startDate, c.src.WorkingDepartureTime)
	if err != nil {
		err = fmt.Errorf("failed to parse WorkingDepartureTime: %w", err)
		return
	}
	databaseLocation.PublicArrivalTime, previousTime, err = convertOptionalTrainTime(previousTime, startDate, c.src.PublicArrivalTime)
	if err != nil {
		err = fmt.Errorf("failed to parse PublicArrivalTime: %w", err)
		return
	}
	databaseLocation.PublicDepartureTime, previousTime, err = convertOptionalTrainTime(previousTime, startDate, c.src.PublicDepartureTime)
	if err != nil {
		err = fmt.Errorf("failed to parse PublicDepartureTime: %w", err)
		return
	}
	databaseLocation.RoutingDelay = parseOptionalRoutingDelay(c.src.RoutingDelay)
	return databaseLocation, previousTime, nil
}

type timetableOperationalDestinationLocation struct {
	log *slog.Logger
	// MUST contain: WTA
	// MAY contain: WTD, RD
	src *unmarshaller.OperationalDestinationTimetableLocation
}

func (c timetableOperationalDestinationLocation) convertToDatabaseLocation(sequence int, previousTime time.Time, startDate time.Time) (databaseLocation repository.ScheduleLocationRow, nextTime time.Time, err error) {
	databaseLocation = newDatabaseLocationWithBaseTimetableValues(c.log, c.src.TimetableLocationBase, sequence)
	databaseLocation.Type = string(unmarshaller.LocationTypeOperationalDestination)
	databaseLocation.WorkingArrivalTime, previousTime, err = convertTrainTime(previousTime, startDate, c.src.WorkingArrivalTime)
	if err != nil {
		err = fmt.Errorf("failed to parse WorkingArrivalTime: %w", err)
		return
	}
	databaseLocation.WorkingDepartureTime, previousTime, err = convertOptionalTrainTime(previousTime, startDate, c.src.WorkingDepartureTime)
	if err != nil {
		err = fmt.Errorf("failed to parse WorkingDepartureTime: %w", err)
		return
	}
	databaseLocation.RoutingDelay = parseOptionalRoutingDelay(c.src.RoutingDelay)
	return databaseLocation, previousTime, nil
}
