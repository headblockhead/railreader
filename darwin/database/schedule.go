package database

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"golang.org/x/net/context"
)

type ScheduleRepository interface {
	Insert(schedule Schedule) error
}

type PGXScheduleRepository struct {
	ctx context.Context
	log *slog.Logger
	tx  pgx.Tx
}

func NewPGXScheduleRepository(ctx context.Context, log *slog.Logger, tx pgx.Tx) PGXScheduleRepository {
	return PGXScheduleRepository{
		ctx: ctx,
		log: log,
		tx:  tx,
	}
}

func (sr PGXScheduleRepository) Insert(s Schedule) error {
	log := sr.log.With(slog.String("schedule_id", s.ScheduleID))
	log.Info("inserting schedule")

	namedArguments := pgx.StrictNamedArgs{
		"message_id":                        s.MessageID,
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

	if _, err := sr.tx.Exec(sr.ctx, `
		INSERT INTO schedules 
			VALUES (
				@schedule_id,
				@message_id,
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
			) ON CONFLICT (schedule_id) DO 
			UPDATE 
				SET
					message_id = EXCLUDED.message_id,
					uid = EXCLUDED.uid,
					scheduled_start_date = EXCLUDED.scheduled_start_date,
					headcode = EXCLUDED.headcode,
					retail_service_id = EXCLUDED.retail_service_id,
					train_operating_company_id = EXCLUDED.train_operating_company_id,
					service = EXCLUDED.service,
					category = EXCLUDED.category,
					is_active = EXCLUDED.is_active,
					is_deleted = EXCLUDED.is_deleted,
					is_charter = EXCLUDED.is_charter,
					cancellation_reason_id = EXCLUDED.cancellation_reason_id,
					cancellation_reason_location_id = EXCLUDED.cancellation_reason_location_id,
					cancellation_reason_is_near_location = EXCLUDED.cancellation_reason_is_near_location,
					late_reason_id = EXCLUDED.late_reason_id,
					late_reason_location_id = EXCLUDED.late_reason_location_id,
					late_reason_is_near_location = EXCLUDED.late_reason_is_near_location,
					diverted_via_location_id = EXCLUDED.diverted_via_location_id;
		`, namedArguments); err != nil {
		return fmt.Errorf("failed to insert schedule %s: %w", s.ScheduleID, err)
	}

	_, err := sr.tx.Exec(sr.ctx, `
		DELETE FROM schedules_locations
			WHERE	schedule_id = @schedule_id;
		`, pgx.NamedArgs{
		"schedule_id": s.ScheduleID,
	})
	if err != pgx.ErrNoRows && err != nil {
		return fmt.Errorf("failed to delete existing schedule for schedule %s: %w", s.ScheduleID, err)
	}

	for _, loc := range s.Locations {
		if err := sr.insertLocation(log.With(slog.Int("sequence", loc.Sequence)), s.ScheduleID, loc); err != nil {
			return fmt.Errorf("failed to process location %s for schedule %s: %w", loc.LocationID, s.ScheduleID, err)
		}
	}

	return nil
}

func (sr PGXScheduleRepository) insertLocation(log *slog.Logger, scheduleID string, location ScheduleLocation) error {
	log.Info("inserting schedule location")
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
	if _, err := sr.tx.Exec(sr.ctx, `
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
	return nil
}

type Schedule struct {
	ScheduleID string

	MessageID string

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
