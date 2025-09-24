package repository

import (
	"context"
	"log/slog"

	"github.com/headblockhead/railreader/database"
	"github.com/jackc/pgx/v5"
)

type LocationRow struct {
	LocationID                      string  `db:"location_id"`
	ComputerisedReservationSystemID *string `db:"computerised_reservation_system_id"`
	TrainOperatingCompanyID         *string `db:"train_operating_company_id"`
	Name                            string  `db:"name"`
}
type Location interface {
	InsertMany(locations []LocationRow) error
}
type PGXLocation struct {
	ctx context.Context
	log *slog.Logger
	tx  pgx.Tx
}

func NewPGXLocation(ctx context.Context, log *slog.Logger, tx pgx.Tx) PGXLocation {
	log.Debug("creating new PGXLocation")
	return PGXLocation{ctx, log, tx}
}

func (lr PGXLocation) InsertMany(locations []LocationRow) error {
	lr.log.Debug("inserting many LocationRows", slog.Int("count", len(locations)))
	return database.InsertManyIntoTable(lr.ctx, lr.tx, "locations", locations)
}

type TrainOperatingCompanyRow struct {
	TrainOperatingCompanyID string  `db:"train_operating_company_id"`
	Name                    string  `db:"name"`
	URL                     *string `db:"url"`
}
type TrainOperatingCompany interface {
	InsertMany(tocs []TrainOperatingCompanyRow) error
}
type PGXTrainOperatingCompany struct {
	ctx context.Context
	log *slog.Logger
	tx  pgx.Tx
}

func NewPGXTrainOperatingCompany(ctx context.Context, log *slog.Logger, tx pgx.Tx) PGXTrainOperatingCompany {
	log.Debug("creating new PGXTrainOperatingCompany")
	return PGXTrainOperatingCompany{ctx, log, tx}
}

func (tr PGXTrainOperatingCompany) InsertMany(tocs []TrainOperatingCompanyRow) error {
	tr.log.Debug("inserting many TrainOperatingCompanyRows", slog.Int("count", len(tocs)))
	return database.InsertManyIntoTable(tr.ctx, tr.tx, "train_operating_companies", tocs)
}

type LateReasonRow struct {
	LateReasonID int    `db:"late_reason_id"`
	Description  string `db:"description"`
}
type LateReason interface {
	InsertMany(lateReasons []LateReasonRow) error
}
type PGXLateReason struct {
	ctx context.Context
	log *slog.Logger
	tx  pgx.Tx
}

func NewPGXLateReason(ctx context.Context, log *slog.Logger, tx pgx.Tx) PGXLateReason {
	log.Debug("creating new PGXLateReason")
	return PGXLateReason{ctx, log, tx}
}

func (lr PGXLateReason) InsertMany(lateReasons []LateReasonRow) error {
	lr.log.Debug("inserting many LateReasonRows", slog.Int("count", len(lateReasons)))
	return database.InsertManyIntoTable(lr.ctx, lr.tx, "late_reasons", lateReasons)
}

type CancellationReasonRow struct {
	CancellationReasonID int    `db:"cancellation_reason_id"`
	Description          string `db:"description"`
}
type CancellationReason interface {
	InsertMany(cancellationReasons []CancellationReasonRow) error
}
type PGXCancellationReason struct {
	ctx context.Context
	log *slog.Logger
	tx  pgx.Tx
}

func NewPGXCancellationReason(ctx context.Context, log *slog.Logger, tx pgx.Tx) PGXCancellationReason {
	log.Debug("creating new PGXCancellationReason")
	return PGXCancellationReason{ctx, log, tx}
}

func (cr PGXCancellationReason) InsertMany(cancellationReasons []CancellationReasonRow) error {
	cr.log.Debug("inserting many CancellationReasonRows", slog.Int("count", len(cancellationReasons)))
	return database.InsertManyIntoTable(cr.ctx, cr.tx, "cancellation_reasons", cancellationReasons)
}

type ViaConditionRow struct {
	Sequence                        int     `db:"sequence"`
	DisplayAtLocationID             string  `db:"display_at_location_id"`
	FirstRequiredCallingLocationID  string  `db:"first_required_calling_location_id"`
	SecondRequiredCallingLocationID *string `db:"second_required_calling_location_id"`
	DestinationRequiredLocationID   string  `db:"destination_required_location_id"`
	Text                            string  `db:"text"`
}
type ViaCondition interface {
	InsertMany(viaConditions []ViaConditionRow) error
}
type PGXViaCondition struct {
	ctx context.Context
	log *slog.Logger
	tx  pgx.Tx
}

func NewPGXViaCondition(ctx context.Context, log *slog.Logger, tx pgx.Tx) PGXViaCondition {
	log.Debug("creating new PGXViaCondition")
	return PGXViaCondition{ctx, log, tx}
}

func (vc PGXViaCondition) InsertMany(viaConditions []ViaConditionRow) error {
	vc.log.Debug("inserting many ViaConditionRows", slog.Int("count", len(viaConditions)))
	return database.InsertManyIntoTable(vc.ctx, vc.tx, "via_conditions", viaConditions)
}

type CustomerInformationSystemRow struct {
	CustomerInformationSystemID string `db:"customer_information_system_id"`
	Name                        string `db:"name"`
}
type CustomerInformationSystem interface {
	InsertMany(cis []CustomerInformationSystemRow) error
}
type PGXCustomerInformationSystem struct {
	ctx context.Context
	log *slog.Logger
	tx  pgx.Tx
}

func NewPGXCustomerInformationSystem(ctx context.Context, log *slog.Logger, tx pgx.Tx) PGXCustomerInformationSystem {
	log.Debug("creating new PGXCustomerInformationSystem")
	return PGXCustomerInformationSystem{ctx, log, tx}
}

func (cis PGXCustomerInformationSystem) InsertMany(cisRows []CustomerInformationSystemRow) error {
	cis.log.Debug("inserting many CustomerInformationSystemRows", slog.Int("count", len(cisRows)))
	return database.InsertManyIntoTable(cis.ctx, cis.tx, "customer_information_systems", cisRows)
}

type LoadingCategoryRow struct {
	// ID is generated by the database.
	LoadingCategoryCode     string  `db:"loading_category_code"`
	TrainOperatingCompanyID *string `db:"train_operating_company_id"`
	Name                    string  `db:"name"`
	DescriptionTypical      string  `db:"description_typical"`
	DescriptionExpected     string  `db:"description_expected"`
	Definition              string  `db:"definition"`
	// can probably ignore these
	Colour string
	Image  string
}
type LoadingCategory interface {
	InsertMany(loadingCategories []LoadingCategoryRow) error
}
type PGXLoadingCategory struct {
	ctx context.Context
	log *slog.Logger
	tx  pgx.Tx
}

func NewPGXLoadingCategory(ctx context.Context, log *slog.Logger, tx pgx.Tx) PGXLoadingCategory {
	log.Debug("creating new PGXLoadingCategory")
	return PGXLoadingCategory{ctx, log, tx}
}

func (lc PGXLoadingCategory) InsertMany(loadingCategories []LoadingCategoryRow) error {
	lc.log.Debug("inserting many LoadingCategoryRows", slog.Int("count", len(loadingCategories)))
	return database.InsertManyIntoTable(lc.ctx, lc.tx, "loading_categories", loadingCategories)
}
