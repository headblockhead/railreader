package interpreter

import (
	"testing"

	"github.com/headblockhead/railreader/darwin/database"
	"github.com/headblockhead/railreader/darwin/unmarshaller"
)

type testingScheduleRepository struct {
	// schedules is a map of ScheduleID to Schedule
	schedules map[string]database.Schedule
}

func (sr *testingScheduleRepository) Insert(schedule database.Schedule) error {
	sr.schedules[schedule.ScheduleID] = schedule
	return nil
}

var interpretScheduleTestCases = map[unmarshaller.Schedule]database.Schedule{
	{}: {},
}

func TestInterpretSchedule(t *testing.T) {
	for input, expected := range interpretScheduleTestCases {
		interpretSchedule(nil, "message1", testingScheduleRepository, input)
	}
}
