package interpreter

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/headblockhead/railreader/darwin/database"
	"github.com/headblockhead/railreader/darwin/unmarshaller"
)

func interpretSchedule(log *slog.Logger, messageID string, lastUpdated time.Time, source string, sourceSystem string, scheduleRepository scheduleRepository, schedule unmarshaller.Schedule) error {
	log.Debug("interpreting a Schedule")
	var databaseSchedule database.Schedule
	databaseSchedule.ScheduleID = schedule.RID
	databaseSchedule.MessageID = messageID
	databaseSchedule.LastUpdated = lastUpdated
	databaseSchedule.Source = source
	databaseSchedule.SourceSystem = sourceSystem
	databaseSchedule.UID = schedule.UID
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
			log.Debug("CancellationReason.TIPLOC is set")
			tiploc := string(schedule.CancellationReason.TIPLOC)
			databaseSchedule.CancellationReasonLocationID = &tiploc
		}
		databaseSchedule.CancellationReasonNearLocation = &schedule.CancellationReason.Near
	}
	if schedule.DiversionReason != nil {
		log.Debug("DiversionReason is set")
		databaseSchedule.LateReasonID = &schedule.DiversionReason.ReasonID
		if schedule.DiversionReason.TIPLOC != "" {
			log.Debug("DiversionReason.TIPLOC is set")
			tiploc := string(schedule.DiversionReason.TIPLOC)
			databaseSchedule.LateReasonLocationID = &tiploc
		}
		databaseSchedule.LateReasonNearLocation = &schedule.DiversionReason.Near
	}

	previousTime := time.Time{}
	for sequence, scheduleLocation := range schedule.Locations {
		locationLog := log.With(slog.Int("sequence", sequence), slog.String("type", string(scheduleLocation.Type)))
		locationLog.Debug("parsing schedule location")
		databaseLocation, nextTime, err := scheduleLocationToDatabaseLocation(locationLog, sequence, startDate, previousTime, scheduleLocation)
		if err != nil {
			return fmt.Errorf("failed to parse schedule location at sequence %d: %w", sequence, err)
		}
		previousTime = nextTime
		databaseSchedule.Locations = append(databaseSchedule.Locations, databaseLocation)
	}

	if err := scheduleRepository.Insert(&databaseSchedule); err != nil {
		return fmt.Errorf("failed to insert schedule %s: %w", schedule.RID, err)
	}
	return nil
}

type databaseableScheduleLocation interface {
	DatabaseLocation(sequence int, previousTime time.Time, startDate time.Time) (databaseLocation database.ScheduleLocation, nextTime time.Time, err error)
}

func scheduleLocationToDatabaseLocation(log *slog.Logger, sequence int, startDate time.Time, previousTime time.Time, scheduleLocation unmarshaller.ScheduleLocation) (databaseLocation database.ScheduleLocation, nextTime time.Time, err error) {
	log.Debug("interpreting location")
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
		return databaseLocation, nextTime, fmt.Errorf("unknown location type %s", scheduleLocation.Type)
	}
	return location.DatabaseLocation(sequence, previousTime, startDate)
}

type scheduleOriginLocation struct {
	log *slog.Logger
	// MUST contain: WTD
	// MAY contain: WTA, PTA, PTD, FD
	src *unmarshaller.OriginLocation
}

func (c scheduleOriginLocation) DatabaseLocation(sequence int, previousTime time.Time, startDate time.Time) (databaseLocation database.ScheduleLocation, nextTime time.Time, err error) {
	databaseLocation = newDatabaseLocationWithBaseValues(c.log, c.src.LocationBase, sequence)
	if c.src.WorkingArrivalTime != "" {
		var wta time.Time
		wta, err = trainTimeToTime(previousTime, c.src.WorkingArrivalTime, startDate)
		if err != nil {
			err = fmt.Errorf("failed to parse WorkingArrivalTime: %w", err)
			return
		}
		databaseLocation.WorkingArrivalTime = &wta
	}
	var wtd time.Time
	wtd, err = trainTimeToTime(previousTime, c.src.WorkingDepartureTime, startDate)
	if err != nil {
		err = fmt.Errorf("failed to parse WorkingDepartureTime: %w", err)
		return
	}
	databaseLocation.WorkingDepartureTime = &wtd
	if c.src.PublicArrivalTime != "" {
		var pta time.Time
		pta, err = trainTimeToTime(previousTime, c.src.PublicArrivalTime, startDate)
		if err != nil {
			err = fmt.Errorf("failed to parse PublicArrivalTime: %w", err)
			return
		}
		databaseLocation.PublicArrivalTime = &pta
	}
	if c.src.PublicDepartureTime != "" {
		var ptd time.Time
		ptd, err = trainTimeToTime(previousTime, c.src.PublicDepartureTime, startDate)
		if err != nil {
			err = fmt.Errorf("failed to parse PublicDepartureTime: %w", err)
			return
		}
		databaseLocation.PublicDepartureTime = &ptd
	}
	if c.src.FalseDestination != "" {
		fd := string(c.src.FalseDestination)
		databaseLocation.FalseDestinationLocationID = &fd
	}

	nextTime = wtd
	return
}

type scheduleOperationalOriginLocation struct {
	log *slog.Logger
	// MUST contain: WTD
	// MAY contain: WTA
	src *unmarshaller.OperationalOriginLocation
}

func (c scheduleOperationalOriginLocation) DatabaseLocation(sequence int, previousTime time.Time, startDate time.Time) (databaseLocation database.ScheduleLocation, nextTime time.Time, err error) {
	databaseLocation = newDatabaseLocationWithBaseValues(c.log, c.src.LocationBase, sequence)
	if c.src.WorkingArrivalTime != "" {
		var wta time.Time
		wta, err = trainTimeToTime(previousTime, c.src.WorkingArrivalTime, startDate)
		if err != nil {
			err = fmt.Errorf("failed to parse WorkingArrivalTime: %w", err)
			return
		}
		databaseLocation.WorkingArrivalTime = &wta
	}
	var wtd time.Time
	wtd, err = trainTimeToTime(previousTime, c.src.WorkingDepartureTime, startDate)
	if err != nil {
		err = fmt.Errorf("failed to parse WorkingDepartureTime: %w", err)
		return
	}
	databaseLocation.WorkingDepartureTime = &wtd
	nextTime = wtd
	return
}

type scheduleIntermediateLocation struct {
	log *slog.Logger
	// MUST contain: WTA, WTD
	// MAY contain: PTA, PTD, FD, RD
	src *unmarshaller.IntermediateLocation
}

func (c scheduleIntermediateLocation) DatabaseLocation(sequence int, previousTime time.Time, startDate time.Time) (databaseLocation database.ScheduleLocation, nextTime time.Time, err error) {
	databaseLocation = newDatabaseLocationWithBaseValues(c.log, c.src.LocationBase, sequence)
	var wta time.Time
	wta, err = trainTimeToTime(previousTime, c.src.WorkingArrivalTime, startDate)
	if err != nil {
		err = fmt.Errorf("failed to parse WorkingArrivalTime: %w", err)
		return
	}
	databaseLocation.WorkingArrivalTime = &wta
	var wtd time.Time
	wtd, err = trainTimeToTime(previousTime, c.src.WorkingDepartureTime, startDate)
	if err != nil {
		err = fmt.Errorf("failed to parse WorkingDepartureTime: %w", err)
		return
	}
	databaseLocation.WorkingDepartureTime = &wtd
	if c.src.PublicArrivalTime != "" {
		var pta time.Time
		pta, err = trainTimeToTime(previousTime, c.src.PublicArrivalTime, startDate)
		if err != nil {
			err = fmt.Errorf("failed to parse PublicArrivalTime: %w", err)
			return
		}
		databaseLocation.PublicArrivalTime = &pta
	}
	if c.src.PublicDepartureTime != "" {
		var ptd time.Time
		ptd, err = trainTimeToTime(previousTime, c.src.PublicDepartureTime, startDate)
		if err != nil {
			err = fmt.Errorf("failed to parse PublicDepartureTime: %w", err)
			return
		}
		databaseLocation.PublicDepartureTime = &ptd
	}
	if c.src.FalseDestination != "" {
		fd := string(c.src.FalseDestination)
		databaseLocation.FalseDestinationLocationID = &fd
	}
	if c.src.RoutingDelay != 0 {
		rd := time.Duration(c.src.RoutingDelay) * time.Minute
		databaseLocation.RoutingDelay = &rd
	}
	nextTime = wtd
	return
}

type scheduleOperationalIntermediateLocation struct {
	log *slog.Logger
	// MUST contain: WTA, WTD
	// MAY contain: RD
	src *unmarshaller.OperationalIntermediateLocation
}

func (c scheduleOperationalIntermediateLocation) DatabaseLocation(sequence int, previousTime time.Time, startDate time.Time) (databaseLocation database.ScheduleLocation, nextTime time.Time, err error) {
	databaseLocation = newDatabaseLocationWithBaseValues(c.log, c.src.LocationBase, sequence)
	var wta time.Time
	wta, err = trainTimeToTime(previousTime, c.src.WorkingArrivalTime, startDate)
	if err != nil {
		err = fmt.Errorf("failed to parse WorkingArrivalTime: %w", err)
		return
	}
	databaseLocation.WorkingArrivalTime = &wta
	var wtd time.Time
	wtd, err = trainTimeToTime(previousTime, c.src.WorkingDepartureTime, startDate)
	if err != nil {
		err = fmt.Errorf("failed to parse WorkingDepartureTime: %w", err)
		return
	}
	databaseLocation.WorkingDepartureTime = &wtd
	if c.src.RoutingDelay != 0 {
		rd := time.Duration(c.src.RoutingDelay) * time.Minute
		databaseLocation.RoutingDelay = &rd
	}
	nextTime = wtd
	return
}

type scheduleIntermediatePassingLocation struct {
	log *slog.Logger
	// MUST contain: WTP
	// MAY contain: RD
	src *unmarshaller.IntermediatePassingLocation
}

func (c scheduleIntermediatePassingLocation) DatabaseLocation(sequence int, previousTime time.Time, startDate time.Time) (databaseLocation database.ScheduleLocation, nextTime time.Time, err error) {
	databaseLocation = newDatabaseLocationWithBaseValues(c.log, c.src.LocationBase, sequence)
	var wtp time.Time
	wtp, err = trainTimeToTime(previousTime, c.src.WorkingPassingTime, startDate)
	if err != nil {
		err = fmt.Errorf("failed to parse WorkingPassingTime: %w", err)
		return
	}
	databaseLocation.WorkingPassingTime = &wtp
	if c.src.RoutingDelay != 0 {
		rd := time.Duration(c.src.RoutingDelay) * time.Minute
		databaseLocation.RoutingDelay = &rd
	}
	nextTime = wtp
	return
}

type scheduleDestinationLocation struct {
	log *slog.Logger
	// MUST contain: WTA
	// MAY contain: WTD, PTA, PTD, RD
	src *unmarshaller.DestinationLocation
}

func (c scheduleDestinationLocation) DatabaseLocation(sequence int, previousTime time.Time, startDate time.Time) (databaseLocation database.ScheduleLocation, nextTime time.Time, err error) {
	databaseLocation = newDatabaseLocationWithBaseValues(c.log, c.src.LocationBase, sequence)
	var wta time.Time
	wta, err = trainTimeToTime(previousTime, c.src.WorkingArrivalTime, startDate)
	if err != nil {
		err = fmt.Errorf("failed to parse WorkingArrivalTime: %w", err)
		return
	}
	databaseLocation.WorkingArrivalTime = &wta
	if c.src.WorkingDepartureTime != "" {
		var wtd time.Time
		wtd, err = trainTimeToTime(previousTime, c.src.WorkingDepartureTime, startDate)
		if err != nil {
			err = fmt.Errorf("failed to parse WorkingDepartureTime: %w", err)
			return
		}
		databaseLocation.WorkingDepartureTime = &wtd
	}
	if c.src.PublicArrivalTime != "" {
		var pta time.Time
		pta, err = trainTimeToTime(previousTime, c.src.PublicArrivalTime, startDate)
		if err != nil {
			err = fmt.Errorf("failed to parse PublicArrivalTime: %w", err)
			return
		}
		databaseLocation.PublicArrivalTime = &pta
	}
	if c.src.PublicDepartureTime != "" {
		var ptd time.Time
		ptd, err = trainTimeToTime(previousTime, c.src.PublicDepartureTime, startDate)
		if err != nil {
			err = fmt.Errorf("failed to parse PublicDepartureTime: %w", err)
			return
		}
		databaseLocation.PublicDepartureTime = &ptd
	}
	if c.src.RoutingDelay != 0 {
		rd := time.Duration(c.src.RoutingDelay) * time.Minute
		databaseLocation.RoutingDelay = &rd
	}
	nextTime = wta
	return
}

type scheduleOperationalDestinationLocation struct {
	log *slog.Logger
	// MUST contain: WTA
	// MAY contain: WTD, RD
	src *unmarshaller.OperationalDestinationLocation
}

func (c scheduleOperationalDestinationLocation) DatabaseLocation(sequence int, previousTime time.Time, startDate time.Time) (databaseLocation database.ScheduleLocation, nextTime time.Time, err error) {
	databaseLocation = newDatabaseLocationWithBaseValues(c.log, c.src.LocationBase, sequence)
	var wta time.Time
	wta, err = trainTimeToTime(previousTime, c.src.WorkingArrivalTime, startDate)
	if err != nil {
		err = fmt.Errorf("failed to parse WorkingArrivalTime: %w", err)
		return
	}
	databaseLocation.WorkingArrivalTime = &wta
	if c.src.WorkingDepartureTime != "" {
		var wtd time.Time
		wtd, err = trainTimeToTime(previousTime, c.src.WorkingDepartureTime, startDate)
		if err != nil {
			err = fmt.Errorf("failed to parse WorkingDepartureTime: %w", err)
			return
		}
		databaseLocation.WorkingDepartureTime = &wtd
	}
	if c.src.RoutingDelay != 0 {
		rd := time.Duration(c.src.RoutingDelay) * time.Minute
		databaseLocation.RoutingDelay = &rd
	}
	nextTime = wta
	return
}

func newDatabaseLocationWithBaseValues(log *slog.Logger, baseValues unmarshaller.LocationBase, sequence int) (databaseLocation database.ScheduleLocation) {
	databaseLocation.Sequence = sequence
	databaseLocation.LocationID = string(baseValues.TIPLOC)
	if baseValues.Activities != "" {
		log.Debug("Activities is set")
		databaseLocation.Activities = &baseValues.Activities
	}
	if baseValues.PlannedActivities != "" {
		log.Debug("PlannedActivities is set")
		databaseLocation.PlannedActivities = &baseValues.PlannedActivities
	}
	databaseLocation.Cancelled = baseValues.Cancelled
	databaseLocation.AffectedByDiversion = baseValues.AffectedByDiversion
	if baseValues.CancellationReason != nil {
		log.Debug("CancellationReason is set")
		databaseLocation.CancellationReasonID = &baseValues.CancellationReason.ReasonID
		if baseValues.CancellationReason.TIPLOC != "" {
			log.Debug("CancellationReason.TIPLOC is set")
			tiploc := string(baseValues.CancellationReason.TIPLOC)
			databaseLocation.CancellationReasonLocationID = &tiploc
		}
		databaseLocation.CancellationReasonNearLocation = &baseValues.CancellationReason.Near
	}
	return
}
