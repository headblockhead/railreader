package interpreter

import (
	"log/slog"

	"github.com/headblockhead/railreader/darwin/database"
)

type UnitOfWork struct {
	log                *slog.Logger
	messageID          string
	scheduleRepository scheduleRepository
}

func NewUnitOfWork(log *slog.Logger, messageID string, scheduleRepository scheduleRepository) UnitOfWork {
	return UnitOfWork{
		log:                log,
		messageID:          messageID,
		scheduleRepository: scheduleRepository,
	}
}

type scheduleRepository interface {
	Insert(schedule *database.Schedule) error
}
