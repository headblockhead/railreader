package repository

import (
	"context"
	"log/slog"

	"github.com/headblockhead/railreader/database"
	"github.com/jackc/pgx/v5"
)

type ReferenceRow struct {
	ReferenceID string `db:"reference_id"`
	Filename    string `db:"filename"`
}
type Reference interface {
	Insert(reference ReferenceRow) error
	SelectLast() (*ReferenceRow, error)
}
type PGXReference struct {
	ctx context.Context
	log *slog.Logger
	tx  pgx.Tx
}

func NewPGXReference(ctx context.Context, log *slog.Logger, tx pgx.Tx) PGXReference {
	return PGXReference{ctx, log, tx}
}
func (r PGXReference) Insert(reference ReferenceRow) error {
	r.log.Debug("inserting ReferenceRow", slog.String("filename_prefix", reference.Filename))
	return database.InsertIntoTable(r.ctx, r.tx, "reference_files", reference)
}
func (r PGXReference) SelectLast() (*ReferenceRow, error) {
	r.log.Debug("getting last ReferenceRow")
	var reference ReferenceRow
	err := r.tx.QueryRow(r.ctx, `
		SELECT (filename) FROM reference_files ORDER BY reference_file_id DESC LIMIT 1;
	`).Scan(&reference.Filename)
	if err != nil {
		return nil, err
	}
	r.log.Debug("fetched last ReferenceRow", slog.String("filename_prefix", reference.Filename))
	return &reference, nil
}

type LocationRow struct {
	LocationID                      string  `db:"location_id"`
	ComputerisedReservationSystemID *string `db:"computerised_reservation_system_id"`
	TrainOperatingCompanyID         *string `db:"train_operating_company_id"`
	Name                            string  `db:"name"`
}
type Location interface {
	InsertMany(locations []LocationRow) error
	DeleteAll() error
}
type PGXLocation struct {
	ctx context.Context
	log *slog.Logger
	tx  pgx.Tx
}

func NewPGXLocation(ctx context.Context, log *slog.Logger, tx pgx.Tx) PGXLocation {
	return PGXLocation{ctx, log, tx}
}

func (lr PGXLocation) InsertMany(locations []LocationRow) error {
	lr.log.Debug("inserting many LocationRows", slog.Int("count", len(locations)))
	return database.InsertManyIntoTable(lr.ctx, lr.tx, "locations", locations)
}
func (lr PGXLocation) DeleteAll() error {
	lr.log.Debug("deleting all LocationRows")
	return database.DeleteAllInTable(lr.ctx, lr.tx, "locations")
}

type TrainOperatingCompanyRow struct {
	TrainOperatingCompanyID string  `db:"train_operating_company_id"`
	Name                    string  `db:"name"`
	URL                     *string `db:"url"`
}
type TrainOperatingCompany interface {
	InsertMany(tocs []TrainOperatingCompanyRow) error
	DeleteAll() error
}
type PGXTrainOperatingCompany struct {
	ctx context.Context
	log *slog.Logger
	tx  pgx.Tx
}

func NewPGXTrainOperatingCompany(ctx context.Context, log *slog.Logger, tx pgx.Tx) PGXTrainOperatingCompany {
	return PGXTrainOperatingCompany{ctx, log, tx}
}

func (tr PGXTrainOperatingCompany) InsertMany(tocs []TrainOperatingCompanyRow) error {
	tr.log.Debug("inserting many TrainOperatingCompanyRows", slog.Int("count", len(tocs)))
	return database.InsertManyIntoTable(tr.ctx, tr.tx, "train_operating_companies", tocs)
}
func (tr PGXTrainOperatingCompany) DeleteAll() error {
	tr.log.Debug("deleting all TrainOperatingCompanyRows")
	return database.DeleteAllInTable(tr.ctx, tr.tx, "train_operating_companies")
}

type LateReasonRow struct {
	LateReasonID int    `db:"late_reason_id"`
	Description  string `db:"description"`
}
type LateReason interface {
	InsertMany(lateReasons []LateReasonRow) error
	DeleteAll() error
}
type PGXLateReason struct {
	ctx context.Context
	log *slog.Logger
	tx  pgx.Tx
}

func NewPGXLateReason(ctx context.Context, log *slog.Logger, tx pgx.Tx) PGXLateReason {
	return PGXLateReason{ctx, log, tx}
}

func (lr PGXLateReason) InsertMany(lateReasons []LateReasonRow) error {
	lr.log.Debug("inserting many LateReasonRows", slog.Int("count", len(lateReasons)))
	return database.InsertManyIntoTable(lr.ctx, lr.tx, "late_reasons", lateReasons)
}
func (lr PGXLateReason) DeleteAll() error {
	lr.log.Debug("deleting all LateReasonRows")
	return database.DeleteAllInTable(lr.ctx, lr.tx, "late_reasons")
}

type CancellationReasonRow struct {
	CancellationReasonID int    `db:"cancellation_reason_id"`
	Description          string `db:"description"`
}
type CancellationReason interface {
	InsertMany(cancellationReasons []CancellationReasonRow) error
	DeleteAll() error
}

type PGXCancellationReason struct {
	ctx context.Context
	log *slog.Logger
	tx  pgx.Tx
}

func NewPGXCancellationReason(ctx context.Context, log *slog.Logger, tx pgx.Tx) PGXCancellationReason {
	return PGXCancellationReason{ctx, log, tx}
}

func (cr PGXCancellationReason) InsertMany(cancellationReasons []CancellationReasonRow) error {
	cr.log.Debug("inserting many CancellationReasonRows", slog.Int("count", len(cancellationReasons)))
	return database.InsertManyIntoTable(cr.ctx, cr.tx, "cancellation_reasons", cancellationReasons)
}
func (cr PGXCancellationReason) DeleteAll() error {
	cr.log.Debug("deleting all CancellationReasonRows")
	return database.DeleteAllInTable(cr.ctx, cr.tx, "cancellation_reasons")
}

type ViaConditionRow struct {
	Sequence                                 int     `db:"sequence"`
	DisplayAtComputerisedReservationSystemID string  `db:"display_at_computerised_reservation_system_id"`
	FirstRequiredLocationID                  string  `db:"first_required_location_id"`
	SecondRequiredLocationID                 *string `db:"second_required_location_id"`
	DestinationRequiredLocationID            string  `db:"destination_required_location_id"`
	Text                                     string  `db:"text"`
}
type ViaCondition interface {
	InsertMany(viaConditions []ViaConditionRow) error
	DeleteAll() error
}
type PGXViaCondition struct {
	ctx context.Context
	log *slog.Logger
	tx  pgx.Tx
}

func NewPGXViaCondition(ctx context.Context, log *slog.Logger, tx pgx.Tx) PGXViaCondition {
	return PGXViaCondition{ctx, log, tx}
}

func (vc PGXViaCondition) InsertMany(viaConditions []ViaConditionRow) error {
	vc.log.Debug("inserting many ViaConditionRows", slog.Int("count", len(viaConditions)))
	return database.InsertManyIntoTable(vc.ctx, vc.tx, "via_conditions", viaConditions)
}
func (vc PGXViaCondition) DeleteAll() error {
	vc.log.Debug("deleting all ViaConditionRows")
	return database.DeleteAllInTable(vc.ctx, vc.tx, "via_conditions")
}

type CustomerInformationSystemRow struct {
	CustomerInformationSystemID string `db:"customer_information_system_id"`
	Name                        string `db:"name"`
}
type CustomerInformationSystem interface {
	InsertMany(cis []CustomerInformationSystemRow) error
	DeleteAll() error
}
type PGXCustomerInformationSystem struct {
	ctx context.Context
	log *slog.Logger
	tx  pgx.Tx
}

func NewPGXCustomerInformationSystem(ctx context.Context, log *slog.Logger, tx pgx.Tx) PGXCustomerInformationSystem {
	return PGXCustomerInformationSystem{ctx, log, tx}
}

func (cis PGXCustomerInformationSystem) InsertMany(cisRows []CustomerInformationSystemRow) error {
	cis.log.Debug("inserting many CustomerInformationSystemRows", slog.Int("count", len(cisRows)))
	for i := range cisRows {
		err := database.InsertIntoTable(cis.ctx, cis.tx, "customer_information_systems", cisRows[i])
		if err != nil {
			return err
		}
	}
	return nil
}
func (cis PGXCustomerInformationSystem) DeleteAll() error {
	cis.log.Debug("deleting all CustomerInformationSystemRows")
	return database.DeleteAllInTable(cis.ctx, cis.tx, "customer_information_systems")
}

type LoadingCategoryRow struct {
	// ID is generated by the database.
	LoadingCategoryCode     string  `db:"loading_category_code"`
	TrainOperatingCompanyID *string `db:"train_operating_company_id"`
	Name                    string  `db:"name"`
	DescriptionTypical      string  `db:"description_typical"`
	DescriptionExpected     string  `db:"description_expected"`
	Definition              string  `db:"definition"`
}
type LoadingCategory interface {
	InsertMany(loadingCategories []LoadingCategoryRow) error
	DeleteAll() error
}
type PGXLoadingCategory struct {
	ctx context.Context
	log *slog.Logger
	tx  pgx.Tx
}

func NewPGXLoadingCategory(ctx context.Context, log *slog.Logger, tx pgx.Tx) PGXLoadingCategory {
	return PGXLoadingCategory{ctx, log, tx}
}

func (lc PGXLoadingCategory) InsertMany(loadingCategories []LoadingCategoryRow) error {
	lc.log.Debug("inserting many LoadingCategoryRows", slog.Int("count", len(loadingCategories)))
	return database.InsertManyIntoTable(lc.ctx, lc.tx, "loading_categories", loadingCategories)
}
func (lc PGXLoadingCategory) DeleteAll() error {
	lc.log.Debug("deleting all LoadingCategoryRows")
	return database.DeleteAllInTable(lc.ctx, lc.tx, "loading_categories")
}
