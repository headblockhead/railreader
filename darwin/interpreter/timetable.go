package interpreter

import (
	"log/slog"

	"github.com/headblockhead/railreader/darwin/repository"
	"github.com/headblockhead/railreader/darwin/unmarshaller"
)

func (u UnitOfWork) InterpretTimetable(log *slog.Logger, scheduleRepository repository.Schedule, associationRepository repository.Association, timetable unmarshaller.Timetable) error {
	log.Debug("interpreting a Timetable")
}
