package darwin

import (
	"log/slog"

	"github.com/headblockhead/railreader/darwin/database"
)

type UnitOfWork struct {
	log       *slog.Logger
	messageID string

	ScheduleRepository ScheduleRepository
}

type ScheduleRepository interface {
	Insert(schedule *database.Schedule) error
}
