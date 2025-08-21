package darwin

import (
	"testing"
	"time"

	"github.com/headblockhead/railreader/darwin/unmarshaller"
)

type trainTimeTestCase struct {
	// Inputs
	StartDate    time.Time
	PreviousTime time.Time
	CurrentTime  unmarshaller.TrainTime

	// Expected outputs
	ExpectedCurrentTime time.Time
}

func newTrainTimeTestCase(t *testing.T, startDate string, previousTime time.Time, currentTime string, expectedCurrentTime time.Time) trainTimeTestCase {
	location, err := time.LoadLocation("Europe/London")
	if err != nil {
		t.Fatalf("Failed to load location: %v", err)
	}
	start, err := time.ParseInLocation("2006-01-02", startDate, location)
	if err != nil {
		t.Fatalf("Failed to parse start date: %v", err)
	}

	return trainTimeTestCase{
		StartDate:           start,
		PreviousTime:        previousTime,
		CurrentTime:         unmarshaller.TrainTime(currentTime),
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
		newTrainTimeTestCase(t, "2025-08-16", time.Date(2025, 8, 16, 23, 59, 30, 0, location), "00:05", time.Date(2025, 8, 17, 0, 5, 0, 0, location)),
		// Time going forwards.
		newTrainTimeTestCase(t, "2025-08-10", time.Date(2025, 8, 10, 8, 0, 0, 0, location), "08:05:30", time.Date(2025, 8, 10, 8, 5, 30, 0, location)),
		// Time going backwards.
		newTrainTimeTestCase(t, "2025-08-10", time.Date(2025, 8, 10, 8, 0, 0, 0, location), "07:58", time.Date(2025, 8, 10, 7, 58, 0, 0, location)),
		// Time crossing midnight backwards.
		// TODO: find an example of this - not sure if this is how it should work.
		newTrainTimeTestCase(t, "2025-08-10", time.Date(2025, 8, 11, 0, 0, 30, 0, location), "23:58", time.Date(2025, 8, 10, 23, 58, 0, 0, location)),
		// No previousTime
		newTrainTimeTestCase(t, "2025-08-10", time.Time{}, "08:00:00", time.Date(2025, 8, 10, 8, 0, 0, 0, location)),
	}
	for _, tc := range testCases {
		time, err := trainTimeToTime(tc.PreviousTime, tc.CurrentTime, tc.StartDate)
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
