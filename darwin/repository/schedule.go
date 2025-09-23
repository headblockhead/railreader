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

type ScheduleLocation interface {
	Insert(schedule ScheduleRow) error
	DeleteBySchedule(scheduleID string) error
}

type PGXScheduleLocation struct {
	ctx context.Context
	log *slog.Logger
	tx  pgx.Tx
}

func NewPGXScheduleLocation(ctx context.Context, log *slog.Logger, tx pgx.Tx) PGXScheduleLocation {
	return PGXScheduleLocation{
		ctx: ctx,
		log: log,
		tx:  tx,
	}
}

func (sr PGXScheduleLocation) Insert(s ScheduleLocationRow) error {
	log := sr.log.With(slog.String("schedule_id", s.ScheduleID), slog.Int("sequence", s.Sequence))
	log.Info("inserting ScheduleLocationRow")
	statement, args := database.BuildInsert("schedules_locations", s)
	if _, err := sr.tx.Exec(sr.ctx, statement, args); err != nil {
		return fmt.Errorf("failed to insert location %d for schedule %s: %w", s.Sequence, s.ScheduleID, err)
	}
	return nil
}

func (sr PGXScheduleLocation) DeleteBySchedule(scheduleID string) error {
}
