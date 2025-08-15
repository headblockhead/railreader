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
	tx, err := c.pgxConnection.Begin(c.context)
	if err != nil {
		return fmt.Errorf("failed to begin transaction while processing a schedule: %w", err)
	}

	// TODO: change logs to make it clear that this is a transaction
	// TODO: possibly move transaction higher to lump schedules and formations together, etc.

	_, err = tx.Exec(c.context, `
		DELETE FROM schedules WHERE schedule_id = @schedule_id;
		`, pgx.NamedArgs{
		"schedule_id": s.ScheduleID,
	})
	if err != pgx.ErrNoRows && err != nil {
		return fmt.Errorf("failed to delete existing schedule for schedule %s: %w", s.ScheduleID, err)
	}
	// schedules_locations rows should CASCADE from the above delete.

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

	if _, err := tx.Exec(c.context, `
		INSERT INTO schedules 
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

	for _, loc := range s.Locations {
		if err := c.insertScheduleLocation(tx, s.ScheduleID, &loc); err != nil {
			return fmt.Errorf("failed to process location %s for schedule %s: %w", loc.LocationID, s.ScheduleID, err)
		}
	}

	if err := tx.Commit(c.context); err != nil {
		return fmt.Errorf("failed to commit transaction while processing a schedule: %w", err)
	}
	return nil
}

func (c *Connection) insertScheduleLocation(tx pgx.Tx, scheduleID string, location *ScheduleLocation) error {
	log := c.log.With(slog.String("schedule_id", scheduleID), slog.Int("sequence", location.Sequence))
	log.Debug("processing schedule location")
	namedArgs := pgx.StrictNamedArgs{
		"schedule_id":                       scheduleID,
		"sequence":                          location.Sequence,
		"location_id":                       location.LocationID,
		"activities":                        location.Activities,
		"planned_activities":                location.PlannedActivities,
		"cancelled":                         location.Cancelled,
		"affected_by_diversion":             location.AffectedByDiversion,
		"type":                              location.Type,
		"public_arrival_time":               location.PublicArrivalTime,
		"public_departure_time":             location.PublicDepartureTime,
		"working_arrival_time":              location.WorkingArrivalTime,
		"working_passing_time":              location.WorkingPassingTime,
		"working_departure_time":            location.WorkingDepartureTime,
		"routing_delay":                     location.RoutingDelay,
		"false_destination_location_id":     location.FalseDestinationLocationID,
		"cancellation_reason_id":            location.CancellationReasonID,
		"cancellation_reason_location_id":   location.CancellationReasonLocationID,
		"cancellation_reason_near_location": location.CancellationReasonNearLocation,
	}
	if _, err := tx.Exec(c.context, `
	INSERT INTO schedules_locations 
		VALUES (
			@schedule_id,
			@sequence,
			@location_id,
			@activities,
			@planned_activities,
			@cancelled,
			@affected_by_diversion,
			@type,
			@public_arrival_time,
			@public_departure_time,
			@working_arrival_time,
			@working_passing_time,
			@working_departure_time,
			@routing_delay,
			@false_destination_location_id,
			@cancellation_reason_id,
			@cancellation_reason_location_id,
			@cancellation_reason_near_location
		);
	`, namedArgs); err != nil {
		return fmt.Errorf("failed to insert schedule location %d of schedule %s: %w", location.Sequence, scheduleID, err)
	}
	log.Info("inserted schedule location")
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
	Sequence int

	LocationID string

	Activities          *string
	PlannedActivities   *string
	Cancelled           bool
	AffectedByDiversion bool

	Type                       string
	PublicArrivalTime          *time.Time
	PublicDepartureTime        *time.Time
	WorkingArrivalTime         *time.Time
	WorkingPassingTime         *time.Time
	WorkingDepartureTime       *time.Time
	RoutingDelay               *time.Duration
	FalseDestinationLocationID *string

	CancellationReasonID           *int
	CancellationReasonLocationID   *string
	CancellationReasonNearLocation *bool
}
