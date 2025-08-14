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

	for _, loc := range schedule.Locations {

	}

	if err := p.databaseConnection.InsertSchedule(&dbs); err != nil {
		return fmt.Errorf("failed to insert schedule %s: %w", schedule.RID, err)
	}
	return nil
}
