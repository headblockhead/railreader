package decoder

import (
	"testing"
	"time"
)

type trainTimeTestCase struct {
	// Inputs
	StartDate    time.Time
	PreviousTime TrainTime
	CurrentTime  TrainTime

	// Expected outputs
	ExpectedCurrentTime time.Time
}

func newTrainTimeTestCase(startDate string, previousTime string, currentTime string, expectedCurrentTime time.Time) trainTimeTestCase {
	location, _ := time.LoadLocation("Europe/London")
	start, _ := time.ParseInLocation("2006-01-02", startDate, location)

	return trainTimeTestCase{
		StartDate:           start,
		PreviousTime:        TrainTime(previousTime),
		CurrentTime:         TrainTime(currentTime),
		ExpectedCurrentTime: expectedCurrentTime,
	}
}

func TestTrainTime(t *testing.T) {
	location, err := time.LoadLocation("Europe/London")
	if err != nil {
		t.Fatalf("Failed to load location: %v", err)
	}
	testCases := []trainTimeTestCase{
		// Time crossing midnight forwards
		newTrainTimeTestCase("2025-08-14", "23:59:00", "00:01:30", time.Date(2025, 8, 15, 0, 1, 30, 0, location)),
		// Time going forwards.
		newTrainTimeTestCase("2025-08-14", "08:00:00", "08:05:30", time.Date(2025, 8, 14, 8, 5, 30, 0, location)),
		// Time going backwards.
		newTrainTimeTestCase("2025-08-14", "08:00:00", "07:58:00", time.Date(2025, 8, 14, 7, 58, 0, 0, location)),
		// Time crossing midnight backwards.
		// TODO: find an example of this - not sure if this is how it should work.
		newTrainTimeTestCase("2025-08-14", "00:00:30", "23:58:00", time.Date(2025, 8, 13, 23, 58, 0, 0, location)),
	}
	for _, tc := range testCases {
		time, err := tc.CurrentTime.Time(tc.PreviousTime, tc.StartDate)
		if err != nil {
			t.Errorf("Error converting TrainTime to time.Time: %v", err)
			continue
		}
		if !(*time).Equal(tc.ExpectedCurrentTime) {
			t.Errorf("Expected %v, got %v for start date %v, previous time %v, current time %v",
				tc.ExpectedCurrentTime, time, tc.StartDate, tc.PreviousTime, tc.CurrentTime)
		}
	}
}
