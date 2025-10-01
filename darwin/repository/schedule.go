package repository

import (
	"log/slog"
	"time"

	"github.com/headblockhead/railreader/database"
	"github.com/jackc/pgx/v5"
	"golang.org/x/net/context"
)

type ScheduleRow struct {
	ScheduleID string `db:"schedule_id"`

	MessageID   *string `db:"message_id"`
	TimetableID *string `db:"timetable_id"`

	UID                     string    `db:"uid"`
	ScheduledStartDate      time.Time `db:"scheduled_start_date"`
	Headcode                string    `db:"headcode"`
	RetailServiceID         *string   `db:"retail_service_id"`
	TrainOperatingCompanyID string    `db:"train_operating_company_id"`
	Service                 string    `db:"service"`
	Category                string    `db:"category"`
	IsPassengerService      bool      `db:"is_passenger_service"`
	IsActive                bool      `db:"is_active"`
	IsDeleted               bool      `db:"is_deleted"`
	IsCharter               bool      `db:"is_charter"`

	CancellationReasonID             *int    `db:"cancellation_reason_id"`
	CancellationReasonLocationID     *string `db:"cancellation_reason_location_id"`
	CancellationReasonIsNearLocation *bool   `db:"cancellation_reason_is_near_location"`

	LateReasonID             *int    `db:"late_reason_id"`
	LateReasonLocationID     *string `db:"late_reason_location_id"`
	LateReasonIsNearLocation *bool   `db:"late_reason_is_near_location"`

	DivertedViaLocationID *string `db:"diverted_via_location_id"`

	IsCancelled bool `db:"is_cancelled"`
}
type Schedule interface {
	Select(scheduleID string) (ScheduleRow, error)
	Insert(schedule ScheduleRow) error
	InsertMany(schedules []ScheduleRow) error
	Update(schedule ScheduleRow) error
	Delete(scheduleID string) error

	Exists(scheduleID string) (bool, error)
}
type PGXSchedule struct {
	ctx context.Context
	log *slog.Logger
	tx  pgx.Tx
}

func NewPGXSchedule(ctx context.Context, log *slog.Logger, tx pgx.Tx) PGXSchedule {
	return PGXSchedule{ctx, log, tx}
}

func (sr PGXSchedule) Select(scheduleID string) (row ScheduleRow, err error) {
	log := sr.log.With(slog.String("schedule_id", scheduleID))
	log.Info("selecting ScheduleRow")
	return database.SelectOneFromTable(sr.ctx, sr.tx, "schedules", ScheduleRow{}, "schedule_id = @schedule_id", pgx.StrictNamedArgs{
		"schedule_id": scheduleID,
	})
}

func (sr PGXSchedule) Insert(s ScheduleRow) error {
	sr.log.Debug("inserting ScheduleRow", slog.String("schedule_id", s.ScheduleID))
	return database.InsertIntoTable(sr.ctx, sr.tx, "schedules", s)
}

func (sr PGXSchedule) InsertMany(schedules []ScheduleRow) error {
	sr.log.Debug("inserting many ScheduleRows", slog.Int("count", len(schedules)))
	return database.InsertManyIntoTable(sr.ctx, sr.tx, "schedules", schedules)
}

func (sr PGXSchedule) Update(s ScheduleRow) error {
	log := sr.log.With(slog.String("schedule_id", s.ScheduleID))
	log.Info("updating ScheduleRow")
	statement, args := database.BuildUpdate("schedules", s, "schedule_id = @schedule_id", pgx.StrictNamedArgs{
		"schedule_id": s.ScheduleID,
	})
	if _, err := sr.tx.Exec(sr.ctx, statement, args); err != nil {
		return err
	}
	return nil
}

func (sr PGXSchedule) Delete(scheduleID string) error {
	log := sr.log.With(slog.String("schedule_id", scheduleID))
	log.Info("deleting ScheduleRow")
	statement, args := database.BuildDelete("schedules", "schedule_id = @schedule_id", pgx.StrictNamedArgs{
		"schedule_id": scheduleID,
	})
	if _, err := sr.tx.Exec(sr.ctx, statement, args); err != nil {
		if err == pgx.ErrNoRows {
			return nil
		}
		return err
	}
	return nil
}

func (sr PGXSchedule) Exists(scheduleID string) (exists bool, err error) {
	log := sr.log.With(slog.String("schedule_id", scheduleID))
	log.Info("checking existence of ScheduleRow")
	err = sr.tx.QueryRow(sr.ctx, `SELECT EXISTS (SELECT 1 FROM schedules WHERE schedule_id = @schedule_id);`, pgx.StrictNamedArgs{
		"schedule_id": scheduleID,
	}).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

type ScheduleLocationRow struct {
	ScheduleID string `db:"schedule_id"`
	Sequence   int    `db:"sequence"`

	LocationID string `db:"location_id"`

	Activities            []string `db:"activities"`
	PlannedActivities     []string `db:"planned_activities"`
	IsCancelled           bool     `db:"is_cancelled"`
	FormationID           *string  `db:"formation_id"`
	IsAffectedByDiversion bool     `db:"is_affected_by_diversion"`

	Type                       string  `db:"type"`
	PublicArrivalTime          *string `db:"public_arrival_time"`
	PublicDepartureTime        *string `db:"public_departure_time"`
	WorkingArrivalTime         *string `db:"working_arrival_time"`
	WorkingPassingTime         *string `db:"working_passing_time"`
	WorkingDepartureTime       *string `db:"working_departure_time"`
	RoutingDelay               *int    `db:"routing_delay"`
	FalseDestinationLocationID *string `db:"false_destination_location_id"`

	CancellationReasonID             *int    `db:"cancellation_reason_id"`
	CancellationReasonLocationID     *string `db:"cancellation_reason_location_id"`
	CancellationReasonIsNearLocation *bool   `db:"cancellation_reason_is_near_location"`

	Platform *string `db:"platform"`
}
type ScheduleLocation interface {
	InsertMany(schedules []ScheduleLocationRow) error
	SelectByScheduleID(scheduleID string) (rows []ScheduleLocationRow, err error)
}
type PGXScheduleLocation struct {
	ctx context.Context
	log *slog.Logger
	tx  pgx.Tx
}

func NewPGXScheduleLocation(ctx context.Context, log *slog.Logger, tx pgx.Tx) PGXScheduleLocation {
	return PGXScheduleLocation{ctx, log, tx}
}

func (sr PGXScheduleLocation) InsertMany(schedules []ScheduleLocationRow) error {
	sr.log.Debug("inserting many ScheduleLocationRows", slog.Int("count", len(schedules)))
	return database.InsertManyIntoTable(sr.ctx, sr.tx, "schedule_locations", schedules)
}
func (sr PGXScheduleLocation) SelectByScheduleID(scheduleID string) (rows []ScheduleLocationRow, err error) {
	sr.log.Info("selecting ScheduleLocationRows by schedule_id", slog.String("schedule_id", scheduleID))
	return database.SelectFromTable(sr.ctx, sr.tx, "schedule_locations", ScheduleLocationRow{}, "schedule_id = @schedule_id", pgx.StrictNamedArgs{
		"schedule_id": scheduleID,
	})
}
