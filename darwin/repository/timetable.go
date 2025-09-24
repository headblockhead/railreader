package repository

import (
	"context"
	"log/slog"
	"time"

	"github.com/headblockhead/railreader/database"
	"github.com/jackc/pgx/v5"
)

type TimetableRow struct {
	TimetableID     string    `db:"timetable_id"`
	FirstReceivedAt time.Time `db:"first_received_at"`
}
type Timetable interface {
	Insert(timetable TimetableRow) error
}
type PGXTimetable struct {
	ctx context.Context
	log *slog.Logger
	tx  pgx.Tx
}

func NewPGXTimetable(ctx context.Context, log *slog.Logger, tx pgx.Tx) PGXTimetable {
	return PGXTimetable{ctx, log, tx}
}

func (mr PGXTimetable) Insert(timetable TimetableRow) error {
	mr.log.Debug("inserting MessageXMLRow", slog.String("timetable_id", timetable.TimetableID))
	return database.InsertIntoTable(mr.ctx, mr.tx, "timetables", timetable)
}
