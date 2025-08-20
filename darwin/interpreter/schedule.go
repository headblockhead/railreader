package interpreter

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/headblockhead/railreader/darwin/db"
	"github.com/headblockhead/railreader/darwin/decoder"
)

func (p *Processor) processSchedule(log *slog.Logger, messageID string, lastUpdated time.Time, source string, sourceSystem string, schedule *decoder.Schedule) error {
	if schedule == nil {
		return errors.New("Schedule is nil")
	}

	scheduleLog := log.With("rid", schedule.RID)
	scheduleLog.Debug("processing Schedule")

	var dbs db.Schedule

	dbs.ScheduleID = schedule.RID
	dbs.MessageID = messageID
	dbs.LastUpdated = lastUpdated
	dbs.Source = source
	dbs.SourceSystem = sourceSystem
	dbs.UID = schedule.UID
	location, err := time.LoadLocation("Europe/London")
	if err != nil {
		return fmt.Errorf("failed to load time location: %w", err)
	}
	startDate, err := time.ParseInLocation("2006-01-02", schedule.ScheduledStartDate, location)
	if err != nil {
		return fmt.Errorf("failed to parse ScheduledStartDate %q: %w", schedule.ScheduledStartDate, err)
	}
	dbs.ScheduledStartDate = startDate
	dbs.Headcode = schedule.Headcode
	if schedule.RetailServiceID != "" {
		scheduleLog.Debug("RetailServiceID is set")
		retailServiceID := schedule.RetailServiceID
		dbs.RetailServiceID = &retailServiceID
	}
	dbs.TrainOperatingCompanyID = string(schedule.TOC)
	dbs.Service = string(schedule.Service)
	dbs.Category = string(schedule.Category)
	dbs.Active = schedule.Active
	dbs.Deleted = schedule.Deleted
	dbs.Charter = schedule.Charter
	if schedule.CancellationReason != nil {
		scheduleLog.Debug("CancellationReason is set")
		dbs.CancellationReasonID = &schedule.CancellationReason.ReasonID
		if schedule.CancellationReason.TIPLOC != "" {
			tiploc := string(schedule.CancellationReason.TIPLOC)
			dbs.CancellationReasonLocationID = &tiploc
		}
		dbs.CancellationReasonNearLocation = &schedule.CancellationReason.Near
	}
	if schedule.DiversionReason != nil {
		scheduleLog.Debug("DiversionReason is set")
		dbs.LateReasonID = &schedule.DiversionReason.ReasonID
		if schedule.DiversionReason.TIPLOC != "" {
			tiploc := string(schedule.DiversionReason.TIPLOC)
			dbs.LateReasonLocationID = &tiploc
		}
		dbs.LateReasonNearLocation = &schedule.DiversionReason.Near
	}

	var previousTime time.Time = time.Time{}

	for seq, loc := range schedule.Locations {
		dbLoc := db.ScheduleLocation{}
		previousTime, err := processScheduleLocation(scheduleLog, seq, startDate, &loc, &dbLoc)
		if err := ; err != nil {
			return fmt.Errorf("failed to process %s location at sequence %d for schedule %s: %w", loc.Type, seq, schedule.RID, err)
		}
		dbs.Locations = append(dbs.Locations, dbLoc)
	}

	if err := p.databaseConnection.InsertSchedule(&dbs); err != nil {
		return fmt.Errorf("failed to insert schedule %s: %w", schedule.RID, err)
	}
	return nil
}

func processScheduleLocation(log *slog.Logger, sequence int, previousTime *time.Time, startDate time.Time, genericLocation *decoder.LocationGeneric, dbLoc *db.ScheduleLocation) error {
	if genericLocation == nil {
		return errors.New("LocationGeneric is nil")
	}

	locationLog := log.With(slog.Int("sequence", sequence), slog.String("type", string(genericLocation.Type)))
	locationLog.Debug("processing Location")

	dbLoc.Sequence = sequence
	dbLoc.Type = string(genericLocation.Type)

	switch genericLocation.Type {
	case decoder.LocationTypeOrigin:
		if err := processScheduleOriginLocation(dbLoc, genericLocation.OriginLocation, previousTime, startDate); err != nil {
			return err
		}
	case decoder.LocationTypeOperationalOrigin:
		if err := processScheduleOperationalOriginLocation(dbLoc, genericLocation.OperationalOriginLocation, previousTime, startDate); err != nil {
			return err
		}
	case decoder.LocationTypeIntermediate:
		if err := processScheduleIntermediateLocation(dbLoc, genericLocation.IntermediateLocation, previousTime, startDate); err != nil {
			return err
		}
	case decoder.LocationTypeOperationalIntermediate:
		if err := processScheduleOperationalIntermediateLocation(dbLoc, genericLocation.OperationalIntermediateLocation, previousTime, startDate); err != nil {
			return err
		}
	case decoder.LocationTypeIntermediatePassing:
		if err := processScheduleIntermediatePassingLocation(dbLoc, genericLocation.IntermediatePassingLocation, previousTime, startDate); err != nil {
			return err
		}
	case decoder.LocationTypeDestination:
		if err := processScheduleDestinationLocation(dbLoc, genericLocation.DestinationLocation, previousTime, startDate); err != nil {
			return err
		}
	case decoder.LocationTypeOperationalDestination:
		if err := processScheduleOperationalDestinationLocation(dbLoc, genericLocation.OperationalDestinationLocation, previousTime, startDate); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown location type %s", genericLocation.Type)
	}

	return nil
}

// MUST contain: WTD
// MAY contain: WTA, PTA, PTD, FD
func processScheduleOriginLocation(dbLoc *db.ScheduleLocation, location *decoder.OriginLocation, previousTime *time.Time, startDate time.Time) error {
	if location == nil {
		return errors.New("location is nil")
	}
	appendSharedValues(dbLoc, location.LocationSchedule)
	if location.WorkingArrivalTime != "" {
		wta, err := trainTimeToTime(*previousTime, location.WorkingArrivalTime, startDate)
		if err != nil {
			return fmt.Errorf("failed to parse WorkingArrivalTime %q: %w", location.WorkingArrivalTime, err)
		}
		dbLoc.WorkingArrivalTime = wta
	}
	wtd, err := trainTimeToTime(*previousTime, location.WorkingDepartureTime, startDate)
	if err != nil {
		return fmt.Errorf("failed to parse WorkingDepartureTime %q: %w", location.WorkingDepartureTime, err)
	}
	dbLoc.WorkingDepartureTime = wtd
	if location.PublicArrivalTime != "" {
		pta, err := trainTimeToTime(*previousTime, location.PublicArrivalTime, startDate)
		if err != nil {
			return fmt.Errorf("failed to parse PublicArrivalTime %q: %w", location.PublicArrivalTime, err)
		}
		dbLoc.PublicArrivalTime = pta
	}
	if location.PublicDepartureTime != "" {
		ptd, err := trainTimeToTime(*previousTime, location.PublicDepartureTime, startDate)
		if err != nil {
			return fmt.Errorf("failed to parse PublicDepartureTime %q: %w", location.PublicDepartureTime, err)
		}
		dbLoc.PublicDepartureTime = ptd
	}
	if location.FalseDestination != "" {
		fd := string(location.FalseDestination)
		dbLoc.FalseDestinationLocationID = &fd
	}
	previousTime = wtd
	return nil
}

// MUST contain: WTD
// MAY contain: WTA
func processScheduleOperationalOriginLocation(dbLoc *db.ScheduleLocation, location *decoder.OperationalOriginLocation, previousTime *time.Time, startDate time.Time) error {
	if location == nil {
		return errors.New("location is nil")
	}
	appendSharedValues(dbLoc, location.LocationSchedule)
	if location.WorkingArrivalTime != "" {
		wta, err := trainTimeToTime(*previousTime, location.WorkingArrivalTime, startDate)
		if err != nil {
			return fmt.Errorf("failed to parse WorkingArrivalTime %q: %w", location.WorkingArrivalTime, err)
		}
		dbLoc.WorkingArrivalTime = wta
	}
	wtd, err := trainTimeToTime(*previousTime, location.WorkingDepartureTime, startDate)
	if err != nil {
		return fmt.Errorf("failed to parse WorkingDepartureTime %q: %w", location.WorkingDepartureTime, err)
	}
	dbLoc.WorkingDepartureTime = wtd
	previousTime = wtd
	return nil
}

// MUST contain: WTA, WTD
// MAY contain: PTA, PTD, FD, RD
func processScheduleIntermediateLocation(dbLoc *db.ScheduleLocation, location *decoder.IntermediateLocation, previousTime *time.Time, startDate time.Time) error {
	if location == nil {
		return errors.New("location is nil")
	}
	appendSharedValues(dbLoc, location.LocationSchedule)
	wta, err := trainTimeToTime(*previousTime, location.WorkingArrivalTime, startDate)
	if err != nil {
		return fmt.Errorf("failed to parse WorkingArrivalTime %q: %w", location.WorkingArrivalTime, err)
	}
	dbLoc.WorkingArrivalTime = wta
	wtd, err := trainTimeToTime(*previousTime, location.WorkingDepartureTime, startDate)
	if err != nil {
		return fmt.Errorf("failed to parse WorkingDepartureTime %q: %w", location.WorkingDepartureTime, err)
	}
	dbLoc.WorkingDepartureTime = wtd
	if location.PublicArrivalTime != "" {
		pta, err := trainTimeToTime(*previousTime, location.PublicArrivalTime, startDate)
		if err != nil {
			return fmt.Errorf("failed to parse PublicArrivalTime %q: %w", location.PublicArrivalTime, err)
		}
		dbLoc.PublicArrivalTime = pta
	}
	if location.PublicDepartureTime != "" {
		ptd, err := trainTimeToTime(*previousTime, location.PublicDepartureTime, startDate)
		if err != nil {
			return fmt.Errorf("failed to parse PublicDepartureTime %q: %w", location.PublicDepartureTime, err)
		}
		dbLoc.PublicDepartureTime = ptd
	}
	if location.FalseDestination != "" {
		fd := string(location.FalseDestination)
		dbLoc.FalseDestinationLocationID = &fd
	}
	if location.RoutingDelay != 0 {
		rd := time.Duration(location.RoutingDelay) * time.Minute
		dbLoc.RoutingDelay = &rd
	}
	previousTime = wtd
	return nil
}

// MUST contain: WTA, WTD
// MAY contain: RD
func processScheduleOperationalIntermediateLocation(dbLoc *db.ScheduleLocation, location *decoder.OperationalIntermediateLocation, previousTime *time.Time, startDate time.Time) error {
	if location == nil {
		return errors.New("location is nil")
	}
	appendSharedValues(dbLoc, location.LocationSchedule)
	wta, err := trainTimeToTime(*previousTime, location.WorkingArrivalTime, startDate)
	if err != nil {
		return fmt.Errorf("failed to parse WorkingArrivalTime %q: %w", location.WorkingArrivalTime, err)
	}
	dbLoc.WorkingArrivalTime = wta
	wtd, err := trainTimeToTime(*previousTime, location.WorkingDepartureTime, startDate)
	if err != nil {
		return fmt.Errorf("failed to parse WorkingDepartureTime %q: %w", location.WorkingDepartureTime, err)
	}
	dbLoc.WorkingDepartureTime = wtd
	if location.RoutingDelay != 0 {
		rd := time.Duration(location.RoutingDelay) * time.Minute
		dbLoc.RoutingDelay = &rd
	}
	previousTime = wtd
	return nil
}

// MUST contain: WTP
// MAY contain: RD
func processScheduleIntermediatePassingLocation(dbLoc *db.ScheduleLocation, location *decoder.IntermediatePassingLocation, previousTime *time.Time, startDate time.Time) error {
	if location == nil {
		return errors.New("location is nil")
	}
	appendSharedValues(dbLoc, location.LocationSchedule)
	wtp, err := trainTimeToTime(*previousTime, location.WorkingPassingTime, startDate)
	if err != nil {
		return fmt.Errorf("failed to parse WorkingPassingTime %q: %w", location.WorkingPassingTime, err)
	}
	dbLoc.WorkingPassingTime = wtp
	if location.RoutingDelay != 0 {
		rd := time.Duration(location.RoutingDelay) * time.Minute
		dbLoc.RoutingDelay = &rd
	}
	previousTime = wtp
	return nil
}

// MUST contain: WTA
// MAY contain: WTD, PTA, PTD, RD
func processScheduleDestinationLocation(dbLoc *db.ScheduleLocation, location *decoder.DestinationLocation, previousTime *time.Time, startDate time.Time) error {
	if location == nil {
		return errors.New("location is nil")
	}
	appendSharedValues(dbLoc, location.LocationSchedule)
	wta, err := trainTimeToTime(*previousTime, location.WorkingArrivalTime, startDate)
	if err != nil {
		return fmt.Errorf("failed to parse WorkingArrivalTime %q: %w", location.WorkingArrivalTime, err)
	}
	dbLoc.WorkingArrivalTime = wta
	if location.WorkingDepartureTime != "" {
		wtd, err := trainTimeToTime(*previousTime, location.WorkingDepartureTime, startDate)
		if err != nil {
			return fmt.Errorf("failed to parse WorkingDepartureTime %q: %w", location.WorkingDepartureTime, err)
		}
		dbLoc.WorkingDepartureTime = wtd
	}
	if location.PublicArrivalTime != "" {
		pta, err := trainTimeToTime(*previousTime, location.PublicArrivalTime, startDate)
		if err != nil {
			return fmt.Errorf("failed to parse PublicArrivalTime %q: %w", location.PublicArrivalTime, err)
		}
		dbLoc.PublicArrivalTime = pta
	}
	if location.PublicDepartureTime != "" {
		ptd, err := trainTimeToTime(*previousTime, location.PublicDepartureTime, startDate)
		if err != nil {
			return fmt.Errorf("failed to parse PublicDepartureTime %q: %w", location.PublicDepartureTime, err)
		}
		dbLoc.PublicDepartureTime = ptd
	}
	if location.RoutingDelay != 0 {
		rd := time.Duration(location.RoutingDelay) * time.Minute
		dbLoc.RoutingDelay = &rd
	}
	previousTime = wta
	return nil
}

// MUST contain: WTA
// MAY contain: WTD, RD
func processScheduleOperationalDestinationLocation(dbLoc *db.ScheduleLocation, location *decoder.OperationalDestinationLocation, previousTime *time.Time, startDate time.Time) error {
	if location == nil {
		return errors.New("location is nil")
	}
	appendSharedValues(dbLoc, location.LocationSchedule)
	wta, err := trainTimeToTime(*previousTime, location.WorkingArrivalTime, startDate)
	if err != nil {
		return fmt.Errorf("failed to parse WorkingArrivalTime %q: %w", location.WorkingArrivalTime, err)
	}
	dbLoc.WorkingArrivalTime = wta
	if location.WorkingDepartureTime != "" {
		wtd, err := trainTimeToTime(*previousTime, location.WorkingDepartureTime, startDate)
		if err != nil {
			return fmt.Errorf("failed to parse WorkingDepartureTime %q: %w", location.WorkingDepartureTime, err)
		}
		dbLoc.WorkingDepartureTime = wtd
	}
	if location.RoutingDelay != 0 {
		rd := time.Duration(location.RoutingDelay) * time.Minute
		dbLoc.RoutingDelay = &rd
	}
	previousTime = wta
	return nil
}

func appendSharedValues(dbLoc *db.ScheduleLocation, locationBase decoder.LocationSchedule) {
	dbLoc.LocationID = string(locationBase.TIPLOC)
	if locationBase.Activities != "" {
		dbLoc.Activities = &locationBase.Activities
	}
	if locationBase.PlannedActivities != "" {
		dbLoc.PlannedActivities = &locationBase.PlannedActivities
	}
	dbLoc.Cancelled = locationBase.Cancelled
	dbLoc.AffectedByDiversion = locationBase.AffectedByDiversion
	if locationBase.CancellationReason != nil {
		dbLoc.CancellationReasonID = &locationBase.CancellationReason.ReasonID
		if locationBase.CancellationReason.TIPLOC != "" {
			tiploc := string(locationBase.CancellationReason.TIPLOC)
			dbLoc.CancellationReasonLocationID = &tiploc
		}
		dbLoc.CancellationReasonNearLocation = &locationBase.CancellationReason.Near
	}
}
