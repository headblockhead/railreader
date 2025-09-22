package repository

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/headblockhead/railreader/database"
	"github.com/jackc/pgx/v5"
	"golang.org/x/net/context"
)

type ScheduleRow struct {
	ScheduleID string `db:"schedule_id"`

	UID                     string    `db:"uid"`
	ScheduledStartDate      time.Time `db:"scheduled_start_date"`
	Headcode                string    `db:"headcode"`
	RetailServiceID         *string   `db:"retail_service_id"`
	TrainOperatingCompanyID string    `db:"train_operating_company_id"`
	Service                 string    `db:"service"`
	Category                string    `db:"category"`
	PassengerService        bool      `db:"is_passenger_service"`
	Active                  bool      `db:"is_active"`
	Deleted                 bool      `db:"is_deleted"`
	Charter                 bool      `db:"is_charter"`

	CancellationReasonID           *int    `db:"cancellation_reason_id"`
	CancellationReasonLocationID   *string `db:"cancellation_reason_location_id"`
	CancellationReasonNearLocation *bool   `db:"cancellation_reason_is_near_location"`

	LateReasonID           *int    `db:"late_reason_id"`
	LateReasonLocationID   *string `db:"late_reason_location_id"`
	LateReasonNearLocation *bool   `db:"late_reason_is_near_location"`

	DivertedViaLocationID *string `db:"diverted_via_location_id"`

	Cancelled bool `db:"is_cancelled"`
}

type ScheduleLocationRow struct {
	ScheduleID string `db:"schedule_id"`
	Sequence   int    `db:"sequence"`

	LocationID string `db:"location_id"`

	Activities          *[]string `db:"activities"`
	PlannedActivities   *[]string `db:"planned_activities"`
	Cancelled           bool      `db:"is_cancelled"`
	FormationID         *string   `db:"formation_id"`
	AffectedByDiversion bool      `db:"is_affected_by_diversion"`

	Type                       string         `db:"type"`
	PublicArrivalTime          *time.Time     `db:"public_arrival_time"`
	PublicDepartureTime        *time.Time     `db:"public_departure_time"`
	WorkingArrivalTime         *time.Time     `db:"working_arrival_time"`
	WorkingPassingTime         *time.Time     `db:"working_passing_time"`
	WorkingDepartureTime       *time.Time     `db:"working_departure_time"`
	RoutingDelay               *time.Duration `db:"routing_delay"`
	FalseDestinationLocationID *string        `db:"false_destination_location_id"`

	CancellationReasonID           *int    `db:"cancellation_reason_id"`
	CancellationReasonLocationID   *string `db:"cancellation_reason_location_id"`
	CancellationReasonNearLocation *bool   `db:"cancellation_reason_is_near_location"`

	Platform *string
}

type Schedule interface {
	Insert(schedule ScheduleRow) error
	Update(schedule ScheduleRow) error
	Select(scheduleID string) (ScheduleRow, error)
	Exists(scheduleID string) (bool, error)
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

func (sr PGXSchedule) Insert(s ScheduleRow) error {
	log := sr.log.With(slog.String("schedule_id", s.ScheduleID))
	log.Info("inserting ScheduleRow")
	statement, args := database.BuildInsert("schedules", s)
	if _, err := sr.tx.Exec(sr.ctx, statement, args); err != nil {
		return fmt.Errorf("failed to insert schedule %s: %w", s.ScheduleID, err)
	}
	return nil
}

func (sr PGXSchedule) Update(s ScheduleRow) error {
	log := sr.log.With(slog.String("schedule_id", s.ScheduleID))
	log.Info("updating ScheduleRow")
	statement, args := database.BuildUpdate("schedules", s, "schedule_id = @schedule_id", pgx.StrictNamedArgs{
		"schedule_id": s.ScheduleID,
	})
	if _, err := sr.tx.Exec(sr.ctx, statement, args); err != nil {
		return fmt.Errorf("failed to update schedule %s: %w", s.ScheduleID, err)
	}
	return nil
}

func (sr PGXSchedule) Select(scheduleID string) (row ScheduleRow, err error) {
	log := sr.log.With(slog.String("schedule_id", scheduleID))
	log.Info("selecting ScheduleRow")
	rows, _ := sr.tx.Query(sr.ctx, "SELECT * FROM schedules WHERE schedule_id = @schedule_id;", pgx.StrictNamedArgs{
		"schedule_id": scheduleID,
	})
	row, err = pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[ScheduleRow])
	if err != nil {
		if err == pgx.ErrNoRows {
			return row, fmt.Errorf("schedule %s not found: %w", scheduleID, err)
		}
		return row, fmt.Errorf("failed to select schedule %s: %w", scheduleID, err)
	}
	return row, nil
}

func (sr PGXSchedule) Exists(scheduleID string) (exists bool, err error) {
	log := sr.log.With(slog.String("schedule_id", scheduleID))
	log.Info("checking existence of ScheduleRow")
	err = sr.tx.QueryRow(sr.ctx, `SELECT EXISTS (SELECT 1 FROM schedules WHERE schedule_id = @schedule_id);`, pgx.StrictNamedArgs{
		"schedule_id": scheduleID,
	}).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check existence of schedule %s: %w", scheduleID, err)
	}
	return exists, nil
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
