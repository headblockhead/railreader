package repositories

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"golang.org/x/net/context"
)

type ScheduleRow struct {
	ScheduleID string

	MessageID string

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

	Locations []ScheduleLocationRow
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
}

type ScheduleRepository interface {
	Insert(schedule ScheduleRow) error
	Select(scheduleID string) (ScheduleRow, error)
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

func (sr PGXScheduleRepository) Select(scheduleID string) (row ScheduleRow, err error) {
	return
}

func (sr PGXScheduleRepository) Insert(s ScheduleRow) error {
	log := sr.log.With(slog.String("schedule_id", s.ScheduleID))
	log.Info("inserting schedule")

	scheduleAlreadyExists := true
	// If it exists, select the current schedule record to compare against
	var existingSchedule ScheduleRow
	row := sr.tx.QueryRow(sr.ctx, `
		SELECT * FROM schedules WHERE schedule_id=@schedule_id
	`, pgx.StrictNamedArgs{
		"schedule_id": s.ScheduleID,
	})
	if err := row.Scan(&existingSchedule); err != nil {
		if err != pgx.ErrNoRows {
			return fmt.Errorf("failed to query existing schedule: %w", err)
		}
		sr.log.Debug("schedule does not already exist")
		scheduleAlreadyExists = false
	}
	var existingSchedulePtr *ScheduleRow = nil
	if scheduleAlreadyExists {
		existingSchedulePtr = &existingSchedule
	}

	if err := sr.generateScheduleMessages(existingSchedulePtr, s); err != nil {
		return fmt.Errorf("failed to generate outgoing messages: %w", err)
	}

	namedArguments := pgx.StrictNamedArgs{
		"message_id":                           s.MessageID,
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
	}

	if _, err := sr.tx.Exec(sr.ctx, `
		INSERT INTO schedules 
			VALUES (
				@schedule_id
				,@message_id
				,@uid
				,@scheduled_start_date
				,@headcode,
				,@retail_service_id,
				,@train_operating_company_id,
				,@service,
				,@category,
				,@is_passenger_service
				,@is_active,
				,@is_deleted
				,@is_charter,
				,@cancellation_reason_id,
				,@cancellation_reason_location_id,
				,@cancellation_reason_is_near_location,
				,@late_reason_id,
				,@late_reason_location_id,
				,@late_reason_is_near_location,
				,@diverted_via_location_id
			) ON CONFLICT (schedule_id) DO 
			UPDATE 
				SET
					message_id = EXCLUDED.message_id
					,uid = EXCLUDED.uid
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
					,diverted_via_location_id = EXCLUDED.diverted_via_location_id;
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

func (sr PGXScheduleRepository) insertLocation(log *slog.Logger, scheduleID string, location ScheduleLocationRow) error {
	log.Info("inserting schedule location")
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
		);
	`, namedArgs); err != nil {
		return fmt.Errorf("failed to insert schedule location %d of schedule %s: %w", location.Sequence, scheduleID, err)
	}
	return nil
}

func (sr PGXScheduleRepository) generateScheduleMessages(previousSchedule *ScheduleRow, currentSchedule ScheduleRow) error {
	/* scheduleAlreadyExists := previousSchedule != nil*/
	/*var activatedTime *time.Time*/
	/*var deactivatedTime *time.Time*/
	/*var deletedTime *time.Time*/
	/*if scheduleAlreadyExists {*/
	/*if !previousSchedule.Active && currentSchedule.Active {*/
	/*activatedTime = &currentSchedule.MessageTime*/
	/*}*/
	/*if previousSchedule.Active && !currentSchedule.Active {*/
	/*deactivatedTime = &currentSchedule.MessageTime*/
	/*}*/
	/*if !previousSchedule.Deleted && currentSchedule.Deleted {*/
	/*deletedTime = &currentSchedule.MessageTime*/
	/*}*/
	/*} else {*/
	/*if currentSchedule.Active {*/
	/*activatedTime = &currentSchedule.MessageTime*/
	/*}*/
	/*}*/
	return nil
}

type ScheduleNewMessage struct {
	ScheduleID string
	Time       time.Time
}

type ScheduleActivationMessage struct {
	ScheduleID string
	Time       time.Time
}
