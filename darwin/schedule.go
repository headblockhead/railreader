package darwin

import (
	"errors"
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
			return fmt.Errorf("failed to process %s location at sequence %d for schedule %s: %w", loc.Type, seq, schedule.RID, err)
		}
		previousTime = nextTime
		databaseSchedule.Locations = append(databaseSchedule.Locations, databaseLocation)
	}

	if err := u.ScheduleRepository.Insert(&databaseSchedule); err != nil {
		return fmt.Errorf("failed to insert schedule %s: %w", schedule.RID, err)
	}
	return nil
}

func genericLocationToDatabaseLocation(log *slog.Logger, sequence int, startDate time.Time, previousTime time.Time, genericLocation *unmarshaller.LocationGeneric) (databaseLocation database.ScheduleLocation, nextTime time.Time, err error) {
	locationLog := log.With(slog.Int("sequence", sequence), slog.String("type", string(genericLocation.Type)))
	locationLog.Debug("processing Location")

	dbLoc.Sequence = sequence
	dbLoc.Type = string(genericLocation.Type)

	switch genericLocation.Type {
	case unmarshaller.LocationTypeOrigin:
		nextTime, err = processScheduleOriginLocation(dbLoc, genericLocation.OriginLocation, previousTime, startDate)
		if err != nil {
			return err
		}
	case unmarshaller.LocationTypeOperationalOrigin:
		if err := processScheduleOperationalOriginLocation(dbLoc, genericLocation.OperationalOriginLocation, previousTime, startDate); err != nil {
			return err
		}
	case unmarshaller.LocationTypeIntermediate:
		if err := processScheduleIntermediateLocation(dbLoc, genericLocation.IntermediateLocation, previousTime, startDate); err != nil {
			return err
		}
	case unmarshaller.LocationTypeOperationalIntermediate:
		if err := processScheduleOperationalIntermediateLocation(dbLoc, genericLocation.OperationalIntermediateLocation, previousTime, startDate); err != nil {
			return err
		}
	case unmarshaller.LocationTypeIntermediatePassing:
		if err := processScheduleIntermediatePassingLocation(dbLoc, genericLocation.IntermediatePassingLocation, previousTime, startDate); err != nil {
			return err
		}
	case unmarshaller.LocationTypeDestination:
		if err := processScheduleDestinationLocation(dbLoc, genericLocation.DestinationLocation, previousTime, startDate); err != nil {
			return err
		}
	case unmarshaller.LocationTypeOperationalDestination:
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
func processScheduleOriginLocation(dbLoc *db.ScheduleLocation, location *unmarshaller.OriginLocation, previousTime *time.Time, startDate time.Time) error {
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
func processScheduleOperationalOriginLocation(dbLoc *db.ScheduleLocation, location *unmarshaller.OperationalOriginLocation, previousTime *time.Time, startDate time.Time) error {
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
func processScheduleIntermediateLocation(dbLoc *db.ScheduleLocation, location *unmarshaller.IntermediateLocation, previousTime *time.Time, startDate time.Time) error {
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
func processScheduleOperationalIntermediateLocation(dbLoc *db.ScheduleLocation, location *unmarshaller.OperationalIntermediateLocation, previousTime *time.Time, startDate time.Time) error {
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
func processScheduleIntermediatePassingLocation(dbLoc *db.ScheduleLocation, location *unmarshaller.IntermediatePassingLocation, previousTime *time.Time, startDate time.Time) error {
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
func processScheduleDestinationLocation(dbLoc *db.ScheduleLocation, location *unmarshaller.DestinationLocation, previousTime *time.Time, startDate time.Time) error {
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
func processScheduleOperationalDestinationLocation(dbLoc *db.ScheduleLocation, location *unmarshaller.OperationalDestinationLocation, previousTime *time.Time, startDate time.Time) error {
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

func appendSharedValues(dbLoc *db.ScheduleLocation, locationBase unmarshaller.LocationSchedule) {
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
