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
	Filename        string    `db:"filename"`
}
type Timetable interface {
	Insert(timetable TimetableRow) error
	SelectLast() (*TimetableRow, error)
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
	mr.log.Debug("inserting TimetableRow", slog.String("filename", timetable.Filename))
	return database.InsertIntoTable(mr.ctx, mr.tx, "timetable_files", timetable)
}

func (mr PGXTimetable) SelectLast() (*TimetableRow, error) {
	mr.log.Debug("getting last TimetableRow")
	var timetable TimetableRow
	err := mr.tx.QueryRow(mr.ctx, `
		SELECT (filename) FROM timetable_files ORDER BY timetable_file_id DESC LIMIT 1;
	`).Scan(&timetable.Filename)
	if err != nil {
		if err == pgx.ErrNoRows {
			mr.log.Debug("no TimetableRow found")
			return nil, nil
		}
		return nil, nil
	}
	mr.log.Debug("fetched last TimetableRow", slog.String("filename", timetable.Filename))
	return &timetable, nil
}
