package db

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
)

// InsertSchedule takes a Schedule struct and creates/updates/deletes records appropriately.
func (c *Connection) InsertSchedule(s *Schedule) error {
	log := c.log.With(slog.String("schedule_id", s.ScheduleID))

	log.Debug("processing schedule")
	tx, err := c.connection.Begin(c.context)
	if err != nil {
		return fmt.Errorf("failed to begin transaction while processing a schedule: %w", err)
	}

	scheduleAlreadyExists := true

	if err := tx.QueryRow(c.context, `
	SELECT schedule_id FROM schedules WHERE schedule_id = @schedule_id;
	`, pgx.NamedArgs{
		"schedule_id": s.ScheduleID,
	}).Scan(nil); err != nil {
		if err != pgx.ErrNoRows {
			return fmt.Errorf("failed to check if schedule %s exists: %w", s.ScheduleID, err)
		}
		log.Debug("schedule does not already exist, will insert a new record")
		scheduleAlreadyExists = false
	}

	// TODO: delete existing schedule locations that are not in the new schedule
	// this is fine because the whole thing is wrapped in a transaction

	for _, loc := range s.Locations {
		if err := c.insertScheduleLocation(tx, s.ScheduleID, &loc); err != nil {
			return fmt.Errorf("failed to process location %s for schedule %s: %w", loc.LocationID, s.ScheduleID, err)
		}
	}

	namedArguments := pgx.StrictNamedArgs{
		"last_updated":                      s.LastUpdated,
		"source":                            s.Source,
		"source_system":                     s.SourceSystem,
		"schedule_id":                       s.ScheduleID,
		"uid":                               s.UID,
		"scheduled_start_date":              s.ScheduledStartDate,
		"headcode":                          s.Headcode,
		"retail_service_id":                 s.RetailServiceID,
		"train_operating_company_id":        s.TrainOperatingCompanyID,
		"service":                           s.Service,
		"category":                          s.Category,
		"active":                            s.Active,
		"deleted":                           s.Deleted,
		"charter":                           s.Charter,
		"cancellation_reason_id":            s.CancellationReasonID,
		"cancellation_reason_location_id":   s.CancellationReasonLocationID,
		"cancellation_reason_near_location": s.CancellationReasonNearLocation,
		"late_reason_id":                    s.LateReasonID,
		"late_reason_location_id":           s.LateReasonLocationID,
		"late_reason_near_location":         s.LateReasonNearLocation,
		"diverted_via_location_id":          s.DivertedViaLocationID,
	}

	if !scheduleAlreadyExists {
		if _, err := tx.Exec(c.context, `
		INSERT INTO public.schedules 
			VALUES (
				@schedule_id,
				@last_updated,
				@source,
				@source_system,
				@uid,
				@scheduled_start_date,
				@headcode, 
				@retail_service_id, 
				@train_operating_company_id, 
				@service, 
				@category, 
				@active, 
				@deleted, 
				@charter, 
				@cancellation_reason_id, 
				@cancellation_reason_location_id, 
				@cancellation_reason_near_location, 
				@late_reason_id, 
				@late_reason_location_id, 
				@late_reason_near_location, 
				@diverted_via_location_id
			);
		`, namedArguments); err != nil {
			return fmt.Errorf("failed to insert schedule %s: %w", s.ScheduleID, err)
		}
		log.Info("inserted schedule")
	} else {
		if _, err := tx.Exec(c.context, `
		UPDATE public.schedules 
			SET 
				last_updated = @last_updated,
				source = @source,
				source_system = @source_system,
				uid = @uid,
				scheduled_start_date = @scheduled_start_date,
				headcode = @headcode,
				retail_service_id = @retail_service_id,
				train_operating_company_id = @train_operating_company_id,
				service = @service,
				category = @category,
				is_active = @active,
				is_deleted = @deleted,
				is_charter = @charter,
				cancellation_reason_id = @cancellation_reason_id,
				cancellation_reason_location_id = @cancellation_reason_location_id,
				cancellation_reason_is_near_location = @cancellation_reason_near_location,
				late_reason_id = @late_reason_id,
				late_reason_location_id = @late_reason_location_id,
				late_reason_is_near_location = @late_reason_near_location,
				diverted_via_location_id = @diverted_via_location_id
			WHERE schedule_id = @schedule_id;
	`, namedArguments); err != nil {
			return fmt.Errorf("failed to update schedule %s: %w", s.ScheduleID, err)
		}
		log.Info("updated schedule")
	}

	if err := tx.Commit(c.context); err != nil {
		return fmt.Errorf("failed to commit transaction while processing a schedule: %w", err)
	}
	return nil
}

func (c *Connection) insertScheduleLocation(tx pgx.Tx, scheduleID string, l *ScheduleLocation) error {
	return nil
}

type Schedule struct {
	ScheduleID string

	LastUpdated  time.Time
	Source       string
	SourceSystem string

	UID                     string
	ScheduledStartDate      time.Time
	Headcode                string
	RetailServiceID         *string
	TrainOperatingCompanyID string
	Service                 string
	Category                string
	Active                  bool
	Deleted                 bool
	Charter                 bool

	CancellationReasonID           *int
	CancellationReasonLocationID   *string
	CancellationReasonNearLocation *bool

	LateReasonID           *int
	LateReasonLocationID   *string
	LateReasonNearLocation *bool

	DivertedViaLocationID *string

	Locations []ScheduleLocation
}

type ScheduleLocation struct {
	// ScheduleID  string
	LocationID string
	Sequence   int

	Activities          string
	PlannedActivities   string
	Cancelled           bool
	AffectedByDiversion bool

	Type                       string
	PublicArrivalTime          time.Time
	PublicDepartureTime        time.Time
	WorkingArrivalTime         time.Time
	WorkingPassingTime         time.Time
	WorkingDepartureTime       time.Time
	RoutingDelay               time.Duration
	FalseDestinationLocationID string

	CancellationReasonID           int
	CancellationReasonLocationID   string
	CancellationReasonNearLocation bool
}
