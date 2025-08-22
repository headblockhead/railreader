package interpreter

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

func TestTrainTime(t *testing.T) {
	location, err := time.LoadLocation("Europe/London")
	if err != nil {
		t.Fatalf("Failed to load location: %v", err)
	}
	testCases := []trainTimeTestCase{
		// Time crossing midnight forwards
		{
			StartDate:           time.Date(2025, 8, 10, 0, 0, 0, 0, location),   // 2025-08-10
			PreviousTime:        time.Date(2025, 8, 10, 23, 55, 0, 0, location), // 2025-08-10 23:55
			CurrentTime:         "00:05",
			ExpectedCurrentTime: time.Date(2025, 8, 11, 0, 5, 0, 0, location), // 2025-08-11 00:05
		},
		// Time going forwards.
		{
			StartDate:           time.Date(2025, 8, 10, 0, 0, 0, 0, location), // 2025-08-10
			PreviousTime:        time.Date(2025, 8, 10, 8, 0, 0, 0, location), // 2025-08-10 08:00
			CurrentTime:         "08:05",
			ExpectedCurrentTime: time.Date(2025, 8, 10, 8, 5, 0, 0, location), // 2025-08-10 08:05
		},
		// Time going backwards.
		{
			StartDate:           time.Date(2025, 8, 10, 0, 0, 0, 0, location), // 2025-08-10
			PreviousTime:        time.Date(2025, 8, 10, 8, 0, 0, 0, location), // 2025-08-10 08:00
			CurrentTime:         "07:55",
			ExpectedCurrentTime: time.Date(2025, 8, 10, 7, 55, 0, 0, location), // 2025-08-10 07:55
		},
		// Time crossing midnight backwards.
		{
			StartDate:           time.Date(2025, 8, 10, 0, 0, 0, 0, location), // 2025-08-10
			PreviousTime:        time.Date(2025, 8, 11, 0, 5, 0, 0, location), // 2025-08-11 00:05
			CurrentTime:         "23:55",
			ExpectedCurrentTime: time.Date(2025, 8, 10, 23, 55, 0, 0, location), // 2025-08-10 23:55
		},
		// No previousTime given
		{
			StartDate:           time.Date(2025, 8, 10, 0, 0, 0, 0, location), // 2025-08-10
			PreviousTime:        time.Time{},                                  // zero
			CurrentTime:         "08:00",
			ExpectedCurrentTime: time.Date(2025, 8, 10, 8, 0, 0, 0, location), // 2025-08-10 08:00
		},
		// Time going forwards after previously crossing midnight
		{
			StartDate:           time.Date(2025, 8, 10, 0, 0, 0, 0, location), // 2025-08-10
			PreviousTime:        time.Date(2025, 8, 11, 8, 0, 0, 0, location), // 2025-08-11 08:00
			CurrentTime:         "08:05",
			ExpectedCurrentTime: time.Date(2025, 8, 11, 8, 5, 0, 0, location), // 2025-08-11 08:05
		},
	}
	for _, tc := range testCases {
		time, err := trainTimeToTime(tc.PreviousTime, tc.CurrentTime, tc.StartDate)
		if err != nil {
			t.Errorf("Error converting TrainTime to time.Time: %v", err)
			continue
		}
		if !time.Equal(tc.ExpectedCurrentTime) {
			t.Errorf("Expected %v, got %v for start date %v, previous time %v, current time %v",
				tc.ExpectedCurrentTime, time, tc.StartDate, tc.PreviousTime, tc.CurrentTime)
		}
	}
}
