package darwin

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/headblockhead/railreader/darwin/db"
	"github.com/headblockhead/railreader/darwin/decoder"
)

func (dc *Connection) processSchedule(log *slog.Logger, s *decoder.Schedule) error {
	log.Debug("processing schedule", slog.String("RID", s.RID))

	var dbs db.Schedule

	dbs.ScheduleID = s.RID
	dbs.UID = s.UID
	startDate, err := time.Parse("2006-01-02", s.ScheduledStartDate)
	if err != nil {
		return fmt.Errorf("failed to parse ScheduledStartDate %q for schedule %s: %w", s.ScheduledStartDate, s.RID, err)
	}
	dbs.ScheduledStartDate = startDate
	dbs.Headcode = s.Headcode
	if s.RetailServiceID != "" {
		dbs.RetailServiceID = &s.RetailServiceID
	}
	dbs.TrainOperatingCompanyID = s.TrainOperatingCompany
	dbs.Service = string(s.Service)
	dbs.Category = string(s.Category)
	dbs.Active = s.Active
	dbs.Deleted = s.Deleted
	dbs.Charter = s.Charter

	// TODO

	if err := dc.databaseConnection.InsertSchedule(&dbs); err != nil {
		return fmt.Errorf("failed to insert schedule %s: %w", s.RID, err)
	}
	return nil
}
