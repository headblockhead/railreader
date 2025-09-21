package repository

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"golang.org/x/net/context"
)

type ScheduleRow struct {
	ScheduleID string

	// one of:
	MessageID   *string
	TimetableID *string

	UID                     string
	ScheduledStartDate      time.Time
	Headcode                string
	RetailServiceID         *string
	TrainOperatingCompanyID string
	Service                 string
	Category                string
	PassengerService        bool
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

	Cancelled bool

	ScheduleLocationRows []ScheduleLocationRow
}

type ScheduleLocationRow struct {
	// ScheduleID  string
	Sequence int

	LocationID string

	Activities          *[]string
	PlannedActivities   *[]string
	Cancelled           bool
	FormationID         *string
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

	Platform *string
}

type Schedule interface {
	Insert(schedule ScheduleRow) error
	Select(scheduleID string) (ScheduleRow, error)
}

type PGXSchedule struct {
	ctx context.Context
	log *slog.Logger
	tx  pgx.Tx
}

func NewPGXSchedule(ctx context.Context, log *slog.Logger, tx pgx.Tx) PGXSchedule {
	return PGXSchedule{
		ctx: ctx,
		log: log,
		tx:  tx,
	}
}

func (sr PGXSchedule) Select(scheduleID string) (row ScheduleRow, err error) {
	// TODO: implement Select on Schedule repository
	return
}

func (sr PGXSchedule) Insert(s ScheduleRow) error {
	log := sr.log.With(slog.String("schedule_id", s.ScheduleID))
	log.Info("inserting ScheduleRow")

	namedArguments := pgx.StrictNamedArgs{
		"schedule_id":                          s.ScheduleID,
		"uid":                                  s.UID,
		"scheduled_start_date":                 s.ScheduledStartDate,
		"headcode":                             s.Headcode,
		"retail_service_id":                    s.RetailServiceID,
		"train_operating_company_id":           s.TrainOperatingCompanyID,
		"service":                              s.Service,
		"category":                             s.Category,
		"is_passenger_service":                 s.PassengerService,
		"is_active":                            s.Active,
		"is_deleted":                           s.Deleted,
		"is_charter":                           s.Charter,
		"cancellation_reason_id":               s.CancellationReasonID,
		"cancellation_reason_location_id":      s.CancellationReasonLocationID,
		"cancellation_reason_is_near_location": s.CancellationReasonNearLocation,
		"late_reason_id":                       s.LateReasonID,
		"late_reason_location_id":              s.LateReasonLocationID,
		"late_reason_is_near_location":         s.LateReasonNearLocation,
		"diverted_via_location_id":             s.DivertedViaLocationID,
		"is_cancelled":                         s.Cancelled,
	}

	if _, err := sr.tx.Exec(sr.ctx, `
		INSERT INTO schedules 
			VALUES (
				@schedule_id
				,@uid
				,@scheduled_start_date
				,@headcode
				,@retail_service_id
				,@train_operating_company_id
				,@service
				,@category
				,@is_passenger_service
				,@is_active
				,@is_deleted
				,@is_charter
				,@cancellation_reason_id
				,@cancellation_reason_location_id
				,@cancellation_reason_is_near_location
				,@late_reason_id
				,@late_reason_location_id
				,@late_reason_is_near_location
				,@diverted_via_location_id
				,@is_cancelled
			) ON CONFLICT (schedule_id) DO 
			UPDATE 
				SET
					uid = EXCLUDED.uid
					,scheduled_start_date = EXCLUDED.scheduled_start_date
					,headcode = EXCLUDED.headcode
					,retail_service_id = EXCLUDED.retail_service_id
					,train_operating_company_id = EXCLUDED.train_operating_company_id
					,service = EXCLUDED.service
					,category = EXCLUDED.category
					,is_passenger_service = EXCLUDED.is_passenger_service
					,is_active = EXCLUDED.is_active
					,is_deleted = EXCLUDED.is_deleted
					,is_charter = EXCLUDED.is_charter
					,cancellation_reason_id = EXCLUDED.cancellation_reason_id
					,cancellation_reason_location_id = EXCLUDED.cancellation_reason_location_id
					,cancellation_reason_is_near_location = EXCLUDED.cancellation_reason_is_near_location
					,late_reason_id = EXCLUDED.late_reason_id
					,late_reason_location_id = EXCLUDED.late_reason_location_id
					,late_reason_is_near_location = EXCLUDED.late_reason_is_near_location
					,diverted_via_location_id = EXCLUDED.diverted_via_location_id
					,is_cancelled = EXCLUDED.is_cancelled
					;
		`, namedArguments); err != nil {
		return fmt.Errorf("failed to insert schedule %s: %w", s.ScheduleID, err)
	}

	if s.MessageID != nil {
		_, err := sr.tx.Exec(sr.ctx, `
		INSERT INTO schedules_messages
			VALUES (
				@message_id
				,@schedule_id
			);
		`, pgx.StrictNamedArgs{
			"schedule_id": s.ScheduleID,
			"message_id":  s.MessageID,
		})
		if err != nil {
			return fmt.Errorf("failed to create schedules_messages entry for schedule %s: %w", s.ScheduleID, err)
		}
	}
	if s.TimetableID != nil {
		_, err := sr.tx.Exec(sr.ctx, `
		INSERT INTO schedules_timetables
			VALUES (
				@timetable_id
				,@schedule_id
			);
		`, pgx.StrictNamedArgs{
			"schedule_id":  s.ScheduleID,
			"timetable_id": s.TimetableID,
		})
		if err != nil {
			return fmt.Errorf("failed to create schedules_timetables entry for schedule %s: %w", s.ScheduleID, err)
		}
	}

	_, err := sr.tx.Exec(sr.ctx, `
		DELETE FROM schedules_locations
			WHERE	schedule_id = @schedule_id;
		`, pgx.StrictNamedArgs{
		"schedule_id": s.ScheduleID,
	})
	if err != pgx.ErrNoRows && err != nil {
		return fmt.Errorf("failed to delete existing schedule locations for schedule %s: %w", s.ScheduleID, err)
	}

	for _, loc := range s.ScheduleLocationRows {
		if err := sr.insertLocation(log.With(slog.Int("sequence", loc.Sequence)), s.ScheduleID, loc); err != nil {
			return fmt.Errorf("failed to process location %s for schedule %s: %w", loc.LocationID, s.ScheduleID, err)
		}
	}

	return nil
}

func (sr PGXSchedule) insertLocation(log *slog.Logger, scheduleID string, location ScheduleLocationRow) error {
	log.Info("inserting ScheduleLocationRow")
	namedArgs := pgx.StrictNamedArgs{
		"schedule_id":                          scheduleID,
		"sequence":                             location.Sequence,
		"location_id":                          location.LocationID,
		"activities":                           location.Activities,
		"planned_activities":                   location.PlannedActivities,
		"is_cancelled":                         location.Cancelled,
		"formation_id":                         location.FormationID,
		"is_affected_by_diversion":             location.AffectedByDiversion,
		"type":                                 location.Type,
		"public_arrival_time":                  location.PublicArrivalTime,
		"public_departure_time":                location.PublicDepartureTime,
		"working_arrival_time":                 location.WorkingArrivalTime,
		"working_passing_time":                 location.WorkingPassingTime,
		"working_departure_time":               location.WorkingDepartureTime,
		"routing_delay":                        location.RoutingDelay,
		"false_destination_location_id":        location.FalseDestinationLocationID,
		"cancellation_reason_id":               location.CancellationReasonID,
		"cancellation_reason_location_id":      location.CancellationReasonLocationID,
		"cancellation_reason_is_near_location": location.CancellationReasonNearLocation,
		"platform":                             location.Platform,
	}
	if _, err := sr.tx.Exec(sr.ctx, `
	INSERT INTO schedules_locations 
		VALUES (
			@schedule_id
			,@sequence
			,@location_id
			,@activities
			,@planned_activities
			,@is_cancelled
			,@formation_id
			,@is_affected_by_diversion
			,@type
			,@public_arrival_time
			,@public_departure_time
			,@working_arrival_time
			,@working_passing_time
			,@working_departure_time
			,@routing_delay
			,@false_destination_location_id
			,@cancellation_reason_id
			,@cancellation_reason_location_id
			,@cancellation_reason_is_near_location
			,@platform
		);
	`, namedArgs); err != nil {
		return fmt.Errorf("failed to insert schedule location %d of schedule %s: %w", location.Sequence, scheduleID, err)
	}
	return nil
}
