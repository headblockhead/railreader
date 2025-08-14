package processor

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/headblockhead/railreader/darwin/db"
	"github.com/headblockhead/railreader/darwin/decoder"
)

func (p *Processor) processSchedule(log *slog.Logger, msg *decoder.PushPortMessage, resp *decoder.Response, schedule *decoder.Schedule) error {
	log.Debug("processing schedule", slog.String("RID", schedule.RID))

	var dbs db.Schedule

	dbs.ScheduleID = schedule.RID
	lastUpdated, err := time.Parse(time.RFC3339Nano, msg.Timestamp)
	if err != nil {
		return fmt.Errorf("failed to parse timestamp %q for schedule %s: %w", msg.Timestamp, schedule.RID, err)
	}
	dbs.LastUpdated = lastUpdated
	dbs.Source = resp.Source
	dbs.SourceSystem = resp.SourceSystem
	dbs.UID = schedule.UID
	location, err := time.LoadLocation("Europe/London")
	if err != nil {
		return fmt.Errorf("failed to load location: %w", err)
	}
	startDate, err := time.ParseInLocation("2006-01-02", schedule.ScheduledStartDate, location)
	if err != nil {
		return fmt.Errorf("failed to parse ScheduledStartDate %q for schedule %s: %w", schedule.ScheduledStartDate, schedule.RID, err)
	}
	dbs.ScheduledStartDate = startDate
	dbs.Headcode = schedule.Headcode
	if schedule.RetailServiceID != "" {
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
		dbs.CancellationReasonID = &schedule.CancellationReason.ReasonID
		if schedule.CancellationReason.TIPLOC != "" {
			tiploc := string(schedule.CancellationReason.TIPLOC)
			dbs.CancellationReasonLocationID = &tiploc
		}
		dbs.CancellationReasonNearLocation = &schedule.CancellationReason.Near
	}
	if schedule.DiversionReason != nil {
		dbs.LateReasonID = &schedule.DiversionReason.ReasonID
		if schedule.DiversionReason.TIPLOC != "" {
			tiploc := string(schedule.DiversionReason.TIPLOC)
			dbs.LateReasonLocationID = &tiploc
		}
		dbs.LateReasonNearLocation = &schedule.DiversionReason.Near
	}

	var previousTime time.Time = time.Time{}

	for seq, loc := range schedule.Locations {
		var dbLoc db.ScheduleLocation
		dbLoc.Sequence = seq

		switch loc.Type {
		case decoder.LocationTypeOrigin:
			location := loc.OriginLocation
			appendSharedValues(&dbLoc, location.LocationSchedule)
			dbLoc.Type = string(decoder.LocationTypeOrigin)
			if location.WorkingArrivalTime != "" {
				wta, err := trainTimeToTime(previousTime, location.WorkingArrivalTime, startDate)
				if err != nil {
					return fmt.Errorf("failed to parse WorkingArrivalTime %q for schedule %s at location sequence %d: %w", location.WorkingArrivalTime, schedule.RID, seq, err)
				}
				dbLoc.WorkingArrivalTime = wta
			}
			wtd, err := trainTimeToTime(previousTime, location.WorkingDepartureTime, startDate)
			if err != nil {
				return fmt.Errorf("failed to parse WorkingDepartureTime %q for schedule %s at location sequence %d: %w", location.WorkingDepartureTime, schedule.RID, seq, err)
			}
			dbLoc.WorkingDepartureTime = wtd
			if location.PublicArrivalTime != "" {
				pta, err := trainTimeToTime(previousTime, location.PublicArrivalTime, startDate)
				if err != nil {
					return fmt.Errorf("failed to parse PublicArrivalTime %q for schedule %s at location sequence %d: %w", location.PublicArrivalTime, schedule.RID, seq, err)
				}
				dbLoc.PublicArrivalTime = pta
			}
			if location.PublicDepartureTime != "" {
				ptd, err := trainTimeToTime(previousTime, location.PublicDepartureTime, startDate)
				if err != nil {
					return fmt.Errorf("failed to parse PublicDepartureTime %q for schedule %s at location sequence %d: %w", location.PublicDepartureTime, schedule.RID, seq, err)
				}
				dbLoc.PublicDepartureTime = ptd
			}
			if location.FalseDestination != "" {
				fd := string(location.FalseDestination)
				dbLoc.FalseDestinationLocationID = &fd
			}
			previousTime = *wtd
		case decoder.LocationTypeOperationalOrigin:
			// TODO: complete other types of location
			appendSharedValues(&dbLoc, loc.OperationalOriginLocation.LocationSchedule)
		case decoder.LocationTypeIntermediate:
			appendSharedValues(&dbLoc, loc.IntermediateLocation.LocationSchedule)
		case decoder.LocationTypeOperationalIntermediate:
			appendSharedValues(&dbLoc, loc.OperationalIntermediateLocation.LocationSchedule)
		case decoder.LocationTypeIntermediatePassing:
			appendSharedValues(&dbLoc, loc.IntermediatePassingLocation.LocationSchedule)
		case decoder.LocationTypeDestination:
			appendSharedValues(&dbLoc, loc.DestinationLocation.LocationSchedule)
		case decoder.LocationTypeOperationalDestination:
			appendSharedValues(&dbLoc, loc.OperationalDestinationLocation.LocationSchedule)
		default:
			return fmt.Errorf("unknown location type %s for schedule %s at sequence %d", loc.Type, schedule.RID, seq)
		}

		dbs.Locations = append(dbs.Locations, dbLoc)
	}

	if err := p.databaseConnection.InsertSchedule(&dbs); err != nil {
		return fmt.Errorf("failed to insert schedule %s: %w", schedule.RID, err)
	}
	return nil
}

func appendSharedValues(location *db.ScheduleLocation, loc decoder.LocationSchedule) {
	location.LocationID = string(loc.TIPLOC)
	if loc.Activities != "" {
		location.Activities = &loc.Activities
	}
	if loc.PlannedActivities != "" {
		location.PlannedActivities = &loc.PlannedActivities
	}
	location.Cancelled = loc.Cancelled
	location.AffectedByDiversion = loc.AffectedByDiversion
	if loc.CancellationReason != nil {
		location.CancellationReasonID = &loc.CancellationReason.ReasonID
		if loc.CancellationReason.TIPLOC != "" {
			tiploc := string(loc.CancellationReason.TIPLOC)
			location.CancellationReasonLocationID = &tiploc
		}
		location.CancellationReasonNearLocation = &loc.CancellationReason.Near
	}
}

func trainTimeToTime(previousTime time.Time, currentTrainTime decoder.TrainTime, date time.Time) (*time.Time, error) {
	location, err := time.LoadLocation("Europe/London")
	if err != nil {
		return nil, fmt.Errorf("failed to load location: %w", err)
	}

	var currentTime time.Time
	if len(currentTrainTime) == 8 {
		currentTime, err = time.ParseInLocation("15:04:05", string(currentTrainTime), location)
		if err != nil {
			return nil, fmt.Errorf("failed to parse time %q: %w", currentTrainTime, err)
		}
	} else if len(currentTrainTime) == 5 {
		currentTime, err = time.ParseInLocation("15:04", string(currentTrainTime), location)
		if err != nil {
			return nil, fmt.Errorf("failed to parse time %q: %w", currentTrainTime, err)
		}
	} else {
		return nil, fmt.Errorf("invalid train time length %q", currentTrainTime)
	}

	if previousTime.IsZero() {
		previousTime = currentTime
	}

	difference := currentTime.Sub(previousTime)

	// Crossed midnight forwards
	if difference < -6*time.Hour {
		date = date.AddDate(0, 0, 1)
	}
	// Backwards time
	if difference < 0 && difference >= -6*time.Hour {
	}
	// Normal time
	if difference >= 0 && difference <= 18*time.Hour {
	}
	// Crossed midnight backwards
	if difference > 18*time.Hour {
		date = date.AddDate(0, 0, -1)
	}

	finalTime := currentTime.AddDate(date.Year(), int(date.Month()-1), date.Day()-1)

	return &finalTime, nil
}
