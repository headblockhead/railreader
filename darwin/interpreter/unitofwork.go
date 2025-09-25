package interpreter

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/headblockhead/railreader/darwin/filegetter"
	"github.com/headblockhead/railreader/darwin/repository"
	"github.com/headblockhead/railreader/database"
	"github.com/jackc/pgx/v5"
)

type UnitOfWork struct {
	ctx       context.Context
	log       *slog.Logger
	messageID string
	tx        pgx.Tx
	fg        filegetter.FileGetter

	// Reference
	locationRepository                  repository.Location
	trainOperatingCompanyRepository     repository.TrainOperatingCompany
	lateReasonRepository                repository.LateReason
	cancellationReasonRepository        repository.CancellationReason
	viaConditionRepository              repository.ViaCondition
	customerInformationSystemRepository repository.CustomerInformationSystem
	loadingCategoryRepository           repository.LoadingCategory

	// MessageXML
	messageXMLRepository repository.MessageXML

	// PPort
	pportMessageRepository repository.PPortMessage
	responseRepository     repository.Response

	// Status
	statusRepository repository.Status

	// Timetable
	timetableRepository repository.Timetable

	// Association
	associationRepository repository.Association

	// Schedule
	scheduleRepository         repository.Schedule
	scheduleLocationRepository repository.ScheduleLocation
}

func NewUnitOfWork(ctx context.Context, log *slog.Logger, messageID string, db database.Database, fg filegetter.FileGetter) (unit UnitOfWork, err error) {
	tx, err := db.BeginTx()
	if err != nil {
		err = fmt.Errorf("failed to begin new transaction: %w", err)
		return
	}
	unit = UnitOfWork{
		ctx,
		log,
		messageID,
		tx,
		fg,

		// Reference
		repository.NewPGXLocation(ctx, log.With(slog.String("repository", "Location")), tx),
		repository.NewPGXTrainOperatingCompany(ctx, log.With(slog.String("repository", "TrainOperatingCompany")), tx),
		repository.NewPGXLateReason(ctx, log.With(slog.String("repository", "LateReason")), tx),
		repository.NewPGXCancellationReason(ctx, log.With(slog.String("repository", "CancellationReason")), tx),
		repository.NewPGXViaCondition(ctx, log.With(slog.String("repository", "ViaCondition")), tx),
		repository.NewPGXCustomerInformationSystem(ctx, log.With(slog.String("repository", "CustomerInformationSystem")), tx),
		repository.NewPGXLoadingCategory(ctx, log.With(slog.String("repository", "LoadingCategory")), tx),

		// MessageXML
		repository.NewPGXMessageXML(ctx, log.With(slog.String("repository", "MessageXML")), tx),

		// PPort
		repository.NewPGXPPortMessage(ctx, log.With(slog.String("repository", "PPortMessage")), tx),
		repository.NewPGXResponse(ctx, log.With(slog.String("repository", "Response")), tx),

		// Status
		repository.NewPGXStatus(ctx, log.With(slog.String("repository", "Status")), tx),

		// Timetable
		repository.NewPGXTimetable(ctx, log.With(slog.String("repository", "Timetable")), tx),

		// Association
		repository.NewPGXAssociation(ctx, log.With(slog.String("repository", "Association")), tx),

		// Schedule
		repository.NewPGXSchedule(ctx, log.With(slog.String("repository", "Schedule")), tx),
		repository.NewPGXScheduleLocation(ctx, log.With(slog.String("repository", "ScheduleLocation")), tx),
	}
	return
}

func (u *UnitOfWork) Commit() error {
	if err := u.tx.Commit(u.ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

func (u *UnitOfWork) Rollback() error {
	if err := u.tx.Rollback(u.ctx); err != nil {
		return fmt.Errorf("failed to rollback transaction: %w", err)
	}
	return nil
}
