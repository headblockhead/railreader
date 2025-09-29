package repository

import (
	"context"
	"log/slog"

	"github.com/headblockhead/railreader/database"
	"github.com/jackc/pgx/v5"
)

type AssociationRow struct {
	Category       string `db:"category"`
	IsCancelled    bool   `db:"is_cancelled"`
	IsDeleted      bool   `db:"is_deleted"`
	MainScheduleID string `db:"main_schedule_id"`
	//MainScheduleLocationSequence       int    `db:"main_schedule_location_sequence"`
	AssociatedScheduleID string `db:"associated_schedule_id"`
	//AssociatedScheduleLocationSequence int    `db:"associated_schedule_location_sequence"`
}

type Association interface {
	Insert(association AssociationRow) error
}
type PGXAssociation struct {
	ctx context.Context
	log *slog.Logger
	tx  pgx.Tx
}

func NewPGXAssociation(ctx context.Context, log *slog.Logger, tx pgx.Tx) PGXAssociation {
	return PGXAssociation{ctx, log, tx}
}

func (ar PGXAssociation) Insert(association AssociationRow) error {
	ar.log.Debug("inserting AssociationRow")
	return database.InsertIntoTable(ar.ctx, ar.tx, "associations", association)
}
