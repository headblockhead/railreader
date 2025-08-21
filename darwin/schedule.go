package darwin

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/headblockhead/railreader/darwin/database"
	"github.com/headblockhead/railreader/darwin/unmarshaller"
)

func (u *UnitOfWork) interpretSchedule(lastUpdated time.Time, source string, sourceSystem string, schedule *unmarshaller.Schedule) error {
	log := u.log.With("rid", schedule.RID)
	log.Debug("interpreting schedule")

	var databaseSchedule database.Schedule

	databaseSchedule.ScheduleID = schedule.RID
	databaseSchedule.MessageID = u.messageID
	databaseSchedule.LastUpdated = lastUpdated
	databaseSchedule.Source = source
	databaseSchedule.SourceSystem = sourceSystem
	databaseSchedule.UID = schedule.UID
	location, err := time.LoadLocation("Europe/London")
	if err != nil {
		return fmt.Errorf("failed to load time location: %w", err)
	}
	startDate, err := time.ParseInLocation("2006-01-02", schedule.ScheduledStartDate, location)
	if err != nil {
		return fmt.Errorf("failed to parse ScheduledStartDate %q: %w", schedule.ScheduledStartDate, err)
	}
	databaseSchedule.ScheduledStartDate = startDate
	databaseSchedule.Headcode = schedule.Headcode
	if schedule.RetailServiceID != "" {
		log.Debug("RetailServiceID is set")
		retailServiceID := schedule.RetailServiceID
		databaseSchedule.RetailServiceID = &retailServiceID
	}
	databaseSchedule.TrainOperatingCompanyID = string(schedule.TOC)
	databaseSchedule.Service = string(schedule.Service)
	databaseSchedule.Category = string(schedule.Category)
	databaseSchedule.Active = schedule.Active
	databaseSchedule.Deleted = schedule.Deleted
	databaseSchedule.Charter = schedule.Charter
	if schedule.CancellationReason != nil {
		log.Debug("CancellationReason is set")
		databaseSchedule.CancellationReasonID = &schedule.CancellationReason.ReasonID
		if schedule.CancellationReason.TIPLOC != "" {
			tiploc := string(schedule.CancellationReason.TIPLOC)
			databaseSchedule.CancellationReasonLocationID = &tiploc
		}
		databaseSchedule.CancellationReasonNearLocation = &schedule.CancellationReason.Near
	}
	if schedule.DiversionReason != nil {
		log.Debug("DiversionReason is set")
		databaseSchedule.LateReasonID = &schedule.DiversionReason.ReasonID
		if schedule.DiversionReason.TIPLOC != "" {
			tiploc := string(schedule.DiversionReason.TIPLOC)
			databaseSchedule.LateReasonLocationID = &tiploc
		}
		databaseSchedule.LateReasonNearLocation = &schedule.DiversionReason.Near
	}

	previousTime := time.Time{}
	for seq, loc := range schedule.Locations {
		databaseLocation, nextTime, err := genericLocationToDatabaseLocation(log, seq, startDate, previousTime, &loc)
		if err != nil {
			return fmt.Errorf("failed to process schedule location at sequence %d: %w", seq, err)
		}
		previousTime = *nextTime
		databaseSchedule.Locations = append(databaseSchedule.Locations, *databaseLocation)
	}

	if err := u.ScheduleRepository.Insert(&databaseSchedule); err != nil {
		return fmt.Errorf("failed to insert schedule %s: %w", schedule.RID, err)
	}
	return nil
}

type databaseableScheduleLocation interface {
	DatabaseLocation(sequence int, previousTime time.Time, startDate time.Time) (databaseLocation *database.ScheduleLocation, nextTime *time.Time, err error)
}

func genericLocationToDatabaseLocation(log *slog.Logger, sequence int, startDate time.Time, previousTime time.Time, genericLocation *unmarshaller.LocationGeneric) (databaseLocation *database.ScheduleLocation, nextTime *time.Time, err error) {
	locationLog := log.With(slog.Int("sequence", sequence))
	locationLog.Debug("processing Location")

	var location databaseableScheduleLocation

	switch genericLocation.Type {
	case unmarshaller.LocationTypeOrigin:
		location = scheduleOriginLocation{log: locationLog, src: genericLocation.Origin}
	case unmarshaller.LocationTypeOperationalOrigin:
		location = scheduleOperationalOriginLocation{log: locationLog, src: genericLocation.OperationalOrigin}
	case unmarshaller.LocationTypeIntermediate:
		location = scheduleIntermediateLocation{log: locationLog, src: genericLocation.Intermediate}
	case unmarshaller.LocationTypeOperationalIntermediate:
		location = scheduleOperationalIntermediateLocation{log: locationLog, src: genericLocation.OperationalIntermediate}
	case unmarshaller.LocationTypeIntermediatePassing:
		location = scheduleIntermediatePassingLocation{log: locationLog, src: genericLocation.IntermediatePassing}
	case unmarshaller.LocationTypeDestination:
		location = scheduleDestinationLocation{log: locationLog, src: genericLocation.Destination}
	case unmarshaller.LocationTypeOperationalDestination:
		location = scheduleOperationalDestinationLocation{log: locationLog, src: genericLocation.OperationalDestination}
	default:
		return nil, nil, fmt.Errorf("unknown location type %s", genericLocation.Type)
	}

	return location.DatabaseLocation(sequence, previousTime, startDate)
}

type scheduleOriginLocation struct {
	log *slog.Logger
	// MUST contain: WTD
	// MAY contain: WTA, PTA, PTD, FD
	src *unmarshaller.OriginLocation
}

func (c scheduleOriginLocation) DatabaseLocation(sequence int, previousTime time.Time, startDate time.Time) (*database.ScheduleLocation, *time.Time, error) {
	databaseLocation := newDatabaseLocationWithBaseValues(c.src.LocationSchedule, sequence)
	if c.src.WorkingArrivalTime != "" {
		wta, err := trainTimeToTime(previousTime, c.src.WorkingArrivalTime, startDate)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse WorkingArrivalTime: %w", err)
		}
		databaseLocation.WorkingArrivalTime = wta
	}
	wtd, err := trainTimeToTime(previousTime, c.src.WorkingDepartureTime, startDate)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse WorkingDepartureTime: %w", err)
	}
	databaseLocation.WorkingDepartureTime = wtd
	if c.src.PublicArrivalTime != "" {
		pta, err := trainTimeToTime(previousTime, c.src.PublicArrivalTime, startDate)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse PublicArrivalTime: %w", err)
		}
		databaseLocation.PublicArrivalTime = pta
	}
	if c.src.PublicDepartureTime != "" {
		ptd, err := trainTimeToTime(previousTime, c.src.PublicDepartureTime, startDate)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse PublicDepartureTime: %w", err)
		}
		databaseLocation.PublicDepartureTime = ptd
	}
	if c.src.FalseDestination != "" {
		fd := string(c.src.FalseDestination)
		databaseLocation.FalseDestinationLocationID = &fd
	}
	return databaseLocation, wtd, nil
}

type scheduleOperationalOriginLocation struct {
	log *slog.Logger
	// MUST contain: WTD
	// MAY contain: WTA
	src *unmarshaller.OperationalOriginLocation
}

func (c scheduleOperationalOriginLocation) DatabaseLocation(sequence int, previousTime time.Time, startDate time.Time) (*database.ScheduleLocation, *time.Time, error) {
	databaseLocation := newDatabaseLocationWithBaseValues(c.src.LocationSchedule, sequence)
	if c.src.WorkingArrivalTime != "" {
		wta, err := trainTimeToTime(previousTime, c.src.WorkingArrivalTime, startDate)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse WorkingArrivalTime: %w", err)
		}
		databaseLocation.WorkingArrivalTime = wta
	}
	wtd, err := trainTimeToTime(previousTime, c.src.WorkingDepartureTime, startDate)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse WorkingDepartureTime: %w", err)
	}
	databaseLocation.WorkingDepartureTime = wtd
	return databaseLocation, wtd, nil
}

type scheduleIntermediateLocation struct {
	log *slog.Logger
	// MUST contain: WTA, WTD
	// MAY contain: PTA, PTD, FD, RD
	src *unmarshaller.IntermediateLocation
}

func (c scheduleIntermediateLocation) DatabaseLocation(sequence int, previousTime time.Time, startDate time.Time) (*database.ScheduleLocation, *time.Time, error) {
	databaseLocation := newDatabaseLocationWithBaseValues(c.src.LocationSchedule, sequence)
	wta, err := trainTimeToTime(previousTime, c.src.WorkingArrivalTime, startDate)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse WorkingArrivalTime: %w", err)
	}
	databaseLocation.WorkingArrivalTime = wta
	wtd, err := trainTimeToTime(previousTime, c.src.WorkingDepartureTime, startDate)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse WorkingDepartureTime: %w", err)
	}
	databaseLocation.WorkingDepartureTime = wtd
	if c.src.PublicArrivalTime != "" {
		pta, err := trainTimeToTime(previousTime, c.src.PublicArrivalTime, startDate)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse PublicArrivalTime: %w", err)
		}
		databaseLocation.PublicArrivalTime = pta
	}
	if c.src.PublicDepartureTime != "" {
		ptd, err := trainTimeToTime(previousTime, c.src.PublicDepartureTime, startDate)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse PublicDepartureTime: %w", err)
		}
		databaseLocation.PublicDepartureTime = ptd
	}
	if c.src.FalseDestination != "" {
		fd := string(c.src.FalseDestination)
		databaseLocation.FalseDestinationLocationID = &fd
	}
	if c.src.RoutingDelay != 0 {
		rd := time.Duration(c.src.RoutingDelay) * time.Minute
		databaseLocation.RoutingDelay = &rd
	}
	return databaseLocation, wtd, nil
}

type scheduleOperationalIntermediateLocation struct {
	log *slog.Logger
	// MUST contain: WTA, WTD
	// MAY contain: RD
	src *unmarshaller.OperationalIntermediateLocation
}

func (c scheduleOperationalIntermediateLocation) DatabaseLocation(sequence int, previousTime time.Time, startDate time.Time) (*database.ScheduleLocation, *time.Time, error) {
	databaseLocation := newDatabaseLocationWithBaseValues(c.src.LocationSchedule, sequence)
	wta, err := trainTimeToTime(previousTime, c.src.WorkingArrivalTime, startDate)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse WorkingArrivalTime: %w", err)
	}
	databaseLocation.WorkingArrivalTime = wta
	wtd, err := trainTimeToTime(previousTime, c.src.WorkingDepartureTime, startDate)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse WorkingDepartureTime: %w", err)
	}
	databaseLocation.WorkingDepartureTime = wtd
	if c.src.RoutingDelay != 0 {
		rd := time.Duration(c.src.RoutingDelay) * time.Minute
		databaseLocation.RoutingDelay = &rd
	}
	return databaseLocation, wtd, nil
}

type scheduleIntermediatePassingLocation struct {
	log *slog.Logger
	// MUST contain: WTP
	// MAY contain: RD
	src *unmarshaller.IntermediatePassingLocation
}

func (c scheduleIntermediatePassingLocation) DatabaseLocation(sequence int, previousTime time.Time, startDate time.Time) (*database.ScheduleLocation, *time.Time, error) {
	databaseLocation := newDatabaseLocationWithBaseValues(c.src.LocationSchedule, sequence)
	wtp, err := trainTimeToTime(previousTime, c.src.WorkingPassingTime, startDate)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse WorkingPassingTime: %w", err)
	}
	databaseLocation.WorkingPassingTime = wtp
	if c.src.RoutingDelay != 0 {
		rd := time.Duration(c.src.RoutingDelay) * time.Minute
		databaseLocation.RoutingDelay = &rd
	}
	return databaseLocation, wtp, nil
}

type scheduleDestinationLocation struct {
	log *slog.Logger
	// MUST contain: WTA
	// MAY contain: WTD, PTA, PTD, RD
	src *unmarshaller.DestinationLocation
}

func (c scheduleDestinationLocation) DatabaseLocation(sequence int, previousTime time.Time, startDate time.Time) (*database.ScheduleLocation, *time.Time, error) {
	databaseLocation := newDatabaseLocationWithBaseValues(c.src.LocationSchedule, sequence)
	wta, err := trainTimeToTime(previousTime, c.src.WorkingArrivalTime, startDate)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse WorkingArrivalTime: %w", err)
	}
	databaseLocation.WorkingArrivalTime = wta
	if c.src.WorkingDepartureTime != "" {
		wtd, err := trainTimeToTime(previousTime, c.src.WorkingDepartureTime, startDate)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse WorkingDepartureTime: %w", err)
		}
		databaseLocation.WorkingDepartureTime = wtd
	}
	if c.src.PublicArrivalTime != "" {
		pta, err := trainTimeToTime(previousTime, c.src.PublicArrivalTime, startDate)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse PublicArrivalTime: %w", err)
		}
		databaseLocation.PublicArrivalTime = pta
	}
	if c.src.PublicDepartureTime != "" {
		ptd, err := trainTimeToTime(previousTime, c.src.PublicDepartureTime, startDate)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse PublicDepartureTime: %w", err)
		}
		databaseLocation.PublicDepartureTime = ptd
	}
	if c.src.RoutingDelay != 0 {
		rd := time.Duration(c.src.RoutingDelay) * time.Minute
		databaseLocation.RoutingDelay = &rd
	}
	return databaseLocation, wta, nil
}

type scheduleOperationalDestinationLocation struct {
	log *slog.Logger
	// MUST contain: WTA
	// MAY contain: WTD, RD
	src *unmarshaller.OperationalDestinationLocation
}

func (c scheduleOperationalDestinationLocation) DatabaseLocation(sequence int, previousTime time.Time, startDate time.Time) (*database.ScheduleLocation, *time.Time, error) {
	databaseLocation := newDatabaseLocationWithBaseValues(c.src.LocationSchedule, sequence)
	wta, err := trainTimeToTime(previousTime, c.src.WorkingArrivalTime, startDate)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse WorkingArrivalTime: %w", err)
	}
	databaseLocation.WorkingArrivalTime = wta
	if c.src.WorkingDepartureTime != "" {
		wtd, err := trainTimeToTime(previousTime, c.src.WorkingDepartureTime, startDate)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse WorkingDepartureTime: %w", err)
		}
		databaseLocation.WorkingDepartureTime = wtd
	}
	if c.src.RoutingDelay != 0 {
		rd := time.Duration(c.src.RoutingDelay) * time.Minute
		databaseLocation.RoutingDelay = &rd
	}
	return databaseLocation, wta, nil
}

func newDatabaseLocationWithBaseValues(baseValues unmarshaller.LocationSchedule, sequence int) (databaseLocation *database.ScheduleLocation) {
	databaseLocation = &database.ScheduleLocation{Sequence: sequence}
	databaseLocation.LocationID = string(baseValues.TIPLOC)
	if baseValues.Activities != "" {
		databaseLocation.Activities = &baseValues.Activities
	}
	if baseValues.PlannedActivities != "" {
		databaseLocation.PlannedActivities = &baseValues.PlannedActivities
	}
	databaseLocation.Cancelled = baseValues.Cancelled
	databaseLocation.AffectedByDiversion = baseValues.AffectedByDiversion
	if baseValues.CancellationReason != nil {
		databaseLocation.CancellationReasonID = &baseValues.CancellationReason.ReasonID
		if baseValues.CancellationReason.TIPLOC != "" {
			tiploc := string(baseValues.CancellationReason.TIPLOC)
			databaseLocation.CancellationReasonLocationID = &tiploc
		}
		databaseLocation.CancellationReasonNearLocation = &baseValues.CancellationReason.Near
	}
	return
}
