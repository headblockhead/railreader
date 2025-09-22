package interpreter

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/headblockhead/railreader"
	"github.com/headblockhead/railreader/darwin/repository"
	"github.com/headblockhead/railreader/darwin/unmarshaller"
)

func interpretSchedule(log *slog.Logger, messageID string, scheduleRepository repository.Schedule, schedule unmarshaller.Schedule) error {
	log.Debug("interpreting a Schedule")
	var row repository.ScheduleRow
	row.ScheduleID = schedule.RID
	row.UID = schedule.UID
	log.Debug("parsing ScheduledStartDate", slog.String("value", schedule.ScheduledStartDate))
	location, err := time.LoadLocation("Europe/London")
	if err != nil {
		return fmt.Errorf("failed to load time location Europe/London: %w", err)
	}
	startDate, err := time.ParseInLocation("2006-01-02", schedule.ScheduledStartDate, location)
	if err != nil {
		return fmt.Errorf("failed to parse ScheduledStartDate %q: %w", schedule.ScheduledStartDate, err)
	}
	log.Debug("parsed ScheduledStartDate", slog.Time("time", startDate))
	row.ScheduledStartDate = startDate
	row.Headcode = schedule.Headcode
	if schedule.RetailServiceID != nil && *schedule.RetailServiceID != "" {
		log.Debug("RetailServiceID is set")
		row.RetailServiceID = schedule.RetailServiceID
	}
	row.TrainOperatingCompanyID = string(schedule.TOC)
	row.Service = string(schedule.Service)
	row.Category = string(schedule.Category)
	row.PassengerService = schedule.PassengerService
	row.Active = schedule.Active
	row.Deleted = schedule.Deleted
	row.Charter = schedule.Charter
	if schedule.CancellationReason != nil {
		log.Debug("CancellationReason is set")
		row.Cancelled = true
		row.CancellationReasonID = &schedule.CancellationReason.ReasonID
		if schedule.CancellationReason.TIPLOC != nil && *schedule.CancellationReason.TIPLOC != "" {
			log.Debug("CancellationReason.TIPLOC is set")
			tiploc := string(*schedule.CancellationReason.TIPLOC)
			row.CancellationReasonLocationID = &tiploc
		}
		row.CancellationReasonNearLocation = &schedule.CancellationReason.Near
	}
	if schedule.DiversionReason != nil {
		log.Debug("DiversionReason is set")
		row.LateReasonID = &schedule.DiversionReason.ReasonID
		if schedule.DiversionReason.TIPLOC != nil && *schedule.DiversionReason.TIPLOC != "" {
			log.Debug("DiversionReason.TIPLOC is set")
			tiploc := string(*schedule.DiversionReason.TIPLOC)
			row.LateReasonLocationID = &tiploc
		}
		row.LateReasonNearLocation = &schedule.DiversionReason.Near
	}

	/* previousTime := time.Time{}*/
	/*previousFormationID := ""*/
	/*for sequence, scheduleLocation := range schedule.Locations {*/
	/*locationLog := log.With(slog.Int("sequence", sequence), slog.String("type", string(scheduleLocation.Type)))*/
	/*locationLog.Debug("parsing schedule location")*/
	/*locationRow, nextTime, nextFormationID, err := convertScheduleLocationToRow(locationLog, sequence, startDate, previousTime, previousFormationID, scheduleLocation)*/
	/*if err != nil {*/
	/*return fmt.Errorf("failed to parse schedule location at sequence %d: %w", sequence, err)*/
	/*}*/
	/*previousTime = nextTime*/
	/*previousFormationID = nextFormationID*/
	/*row.ScheduleLocationRows = append(row.ScheduleLocationRows, locationRow)*/
	/*}*/

	return scheduleRepository.Insert(row)
}

type databaseableScheduleLocation interface {
	convertToDatabaseLocation(sequence int, previousTime time.Time, startDate time.Time) (databaseLocation repository.ScheduleLocationRow, nextTime time.Time, err error)
}

func convertScheduleLocationToRow(log *slog.Logger, sequence int, startDate time.Time, previousTime time.Time, previousFormationID string, scheduleLocation unmarshaller.ScheduleLocation) (databaseLocation repository.ScheduleLocationRow, nextTime time.Time, nextFormationID string, err error) {
	log.Debug("converting location")
	var location databaseableScheduleLocation
	switch scheduleLocation.Type {
	case unmarshaller.LocationTypeOrigin:
		location = scheduleOriginLocation{log: log, src: scheduleLocation.Origin}
	case unmarshaller.LocationTypeOperationalOrigin:
		location = scheduleOperationalOriginLocation{log: log, src: scheduleLocation.OperationalOrigin}
	case unmarshaller.LocationTypeIntermediate:
		location = scheduleIntermediateLocation{log: log, src: scheduleLocation.Intermediate}
	case unmarshaller.LocationTypeOperationalIntermediate:
		location = scheduleOperationalIntermediateLocation{log: log, src: scheduleLocation.OperationalIntermediate}
	case unmarshaller.LocationTypeIntermediatePassing:
		location = scheduleIntermediatePassingLocation{log: log, src: scheduleLocation.IntermediatePassing}
	case unmarshaller.LocationTypeDestination:
		location = scheduleDestinationLocation{log: log, src: scheduleLocation.Destination}
	case unmarshaller.LocationTypeOperationalDestination:
		location = scheduleOperationalDestinationLocation{log: log, src: scheduleLocation.OperationalDestination}
	default:
		err = fmt.Errorf("unknown location type %s", scheduleLocation.Type)
		return
	}
	databaseLocation, nextTime, err = location.convertToDatabaseLocation(sequence, previousTime, startDate)
	if err != nil {
		err = fmt.Errorf("failed to convert location to database location: %w", err)
		return
	}
	if databaseLocation.Cancelled {
		log.Debug("location is marked as cancelled")
		nextFormationID = ""
		return
	}
	if databaseLocation.FormationID == nil && previousFormationID != "" {
		log.Debug("carrying forward previous FormationID", slog.String("FormationID", previousFormationID))
		databaseLocation.FormationID = &previousFormationID
		return
	}
	if databaseLocation.FormationID != nil {
		log.Debug("updating previous FormationID", slog.String("FormationID", *databaseLocation.FormationID))
		nextFormationID = *databaseLocation.FormationID
	}
	return
}

func activitiesToSlice(activities string) (result []string) {
	if activities == "" {
		result = append(result, "  ")
		return
	}
	for i := 0; i < len(activities); i += 2 {
		if i+2 <= len(activities) {
			result = append(result, activities[i:i+2])
		} else {
			result = append(result, activities[i:])
		}
	}
	return
}

func newDatabaseLocationWithBaseValues(log *slog.Logger, baseValues unmarshaller.LocationBase, sequence int) (databaseLocation repository.ScheduleLocationRow) {
	databaseLocation.Sequence = sequence
	databaseLocation.LocationID = string(baseValues.TIPLOC)
	if baseValues.Activities != nil {
		log.Debug("Activities is set")
		activities := activitiesToSlice(*baseValues.Activities)
		databaseLocation.Activities = &activities
	}
	if baseValues.PlannedActivities != nil {
		log.Debug("PlannedActivities is set")
		plannedActivities := activitiesToSlice(*baseValues.PlannedActivities)
		databaseLocation.PlannedActivities = &plannedActivities
	}
	if baseValues.FormationID != nil && *baseValues.FormationID != "" {
		log.Debug("FormationID is set")
		databaseLocation.FormationID = baseValues.FormationID
	}
	databaseLocation.Cancelled = baseValues.Cancelled
	databaseLocation.AffectedByDiversion = baseValues.AffectedByDiversion
	if baseValues.CancellationReason != nil {
		log.Debug("CancellationReason is set")
		databaseLocation.CancellationReasonID = &baseValues.CancellationReason.ReasonID
		if baseValues.CancellationReason.TIPLOC != nil && *baseValues.CancellationReason.TIPLOC != "" {
			log.Debug("CancellationReason.TIPLOC is set")
			tiploc := string(*baseValues.CancellationReason.TIPLOC)
			databaseLocation.CancellationReasonLocationID = &tiploc
		}
		databaseLocation.CancellationReasonNearLocation = &baseValues.CancellationReason.Near
	}
	return
}

func convertOptionalTrainTime(previousTime time.Time, startDate time.Time, trainTime *unmarshaller.TrainTime) (convertedTime *time.Time, nextTime time.Time, err error) {
	if trainTime != nil && *trainTime != "" {
		converted, nextTime, err := convertTrainTime(previousTime, startDate, *trainTime)
		if err != nil {
			return nil, previousTime, err
		}
		return converted, nextTime, nil
	}
	return nil, previousTime, nil
}

func convertTrainTime(previousTime time.Time, startDate time.Time, trainTime unmarshaller.TrainTime) (convertedTime *time.Time, nextTime time.Time, err error) {
	converted, err := trainTimeToTime(previousTime, trainTime, startDate)
	if err != nil {
		return nil, previousTime, err
	}
	return &converted, converted, nil
}

func parseOptionalFalseDestination(falseDestination *railreader.TimingPointLocationCode) *string {
	if falseDestination != nil && *falseDestination != "" {
		fd := string(*falseDestination)
		return &fd
	}
	return nil
}

func parseOptionalRoutingDelay(routingDelay *int) *time.Duration {
	if routingDelay != nil && *routingDelay != 0 {
		rd := time.Duration(*routingDelay) * time.Minute
		return &rd
	}
	return nil
}

type scheduleOriginLocation struct {
	log *slog.Logger
	// MUST contain: WTD
	// MAY contain: WTA, PTA, PTD, FD
	src *unmarshaller.OriginLocation
}

func (c scheduleOriginLocation) convertToDatabaseLocation(sequence int, previousTime time.Time, startDate time.Time) (databaseLocation repository.ScheduleLocationRow, nextTime time.Time, err error) {
	databaseLocation = newDatabaseLocationWithBaseValues(c.log, c.src.LocationBase, sequence)
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

type scheduleOperationalOriginLocation struct {
	log *slog.Logger
	// MUST contain: WTD
	// MAY contain: WTA
	src *unmarshaller.OperationalOriginLocation
}

func (c scheduleOperationalOriginLocation) convertToDatabaseLocation(sequence int, previousTime time.Time, startDate time.Time) (databaseLocation repository.ScheduleLocationRow, nextTime time.Time, err error) {
	databaseLocation = newDatabaseLocationWithBaseValues(c.log, c.src.LocationBase, sequence)
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

type scheduleIntermediateLocation struct {
	log *slog.Logger
	// MUST contain: WTA, WTD
	// MAY contain: PTA, PTD, FD, RD
	src *unmarshaller.IntermediateLocation
}

func (c scheduleIntermediateLocation) convertToDatabaseLocation(sequence int, previousTime time.Time, startDate time.Time) (databaseLocation repository.ScheduleLocationRow, nextTime time.Time, err error) {
	databaseLocation = newDatabaseLocationWithBaseValues(c.log, c.src.LocationBase, sequence)
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

type scheduleOperationalIntermediateLocation struct {
	log *slog.Logger
	// MUST contain: WTA, WTD
	// MAY contain: RD
	src *unmarshaller.OperationalIntermediateLocation
}

func (c scheduleOperationalIntermediateLocation) convertToDatabaseLocation(sequence int, previousTime time.Time, startDate time.Time) (databaseLocation repository.ScheduleLocationRow, nextTime time.Time, err error) {
	databaseLocation = newDatabaseLocationWithBaseValues(c.log, c.src.LocationBase, sequence)
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

type scheduleIntermediatePassingLocation struct {
	log *slog.Logger
	// MUST contain: WTP
	// MAY contain: RD
	src *unmarshaller.IntermediatePassingLocation
}

func (c scheduleIntermediatePassingLocation) convertToDatabaseLocation(sequence int, previousTime time.Time, startDate time.Time) (databaseLocation repository.ScheduleLocationRow, nextTime time.Time, err error) {
	databaseLocation = newDatabaseLocationWithBaseValues(c.log, c.src.LocationBase, sequence)
	databaseLocation.Type = string(unmarshaller.LocationTypeIntermediatePassing)
	databaseLocation.WorkingPassingTime, previousTime, err = convertTrainTime(previousTime, startDate, c.src.WorkingPassingTime)
	if err != nil {
		err = fmt.Errorf("failed to parse WorkingPassingTime: %w", err)
		return
	}
	databaseLocation.RoutingDelay = parseOptionalRoutingDelay(c.src.RoutingDelay)
	return databaseLocation, previousTime, nil
}

type scheduleDestinationLocation struct {
	log *slog.Logger
	// MUST contain: WTA
	// MAY contain: WTD, PTA, PTD, RD
	src *unmarshaller.DestinationLocation
}

func (c scheduleDestinationLocation) convertToDatabaseLocation(sequence int, previousTime time.Time, startDate time.Time) (databaseLocation repository.ScheduleLocationRow, nextTime time.Time, err error) {
	databaseLocation = newDatabaseLocationWithBaseValues(c.log, c.src.LocationBase, sequence)
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

type scheduleOperationalDestinationLocation struct {
	log *slog.Logger
	// MUST contain: WTA
	// MAY contain: WTD, RD
	src *unmarshaller.OperationalDestinationLocation
}

func (c scheduleOperationalDestinationLocation) convertToDatabaseLocation(sequence int, previousTime time.Time, startDate time.Time) (databaseLocation repository.ScheduleLocationRow, nextTime time.Time, err error) {
	databaseLocation = newDatabaseLocationWithBaseValues(c.log, c.src.LocationBase, sequence)
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
