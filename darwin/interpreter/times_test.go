package interpreter

import (
	"testing"
	"time"

	"github.com/headblockhead/railreader/darwin/unmarshaller"
)

func TestTrainTime(t *testing.T) {
	location, err := time.LoadLocation("Europe/London")
	if err != nil {
		t.Fatalf("Failed to load location: %v", err)
	}
	testCases := map[string]struct {
		// Inputs
		StartDate    time.Time
		PreviousTime time.Time
		CurrentTime  unmarshaller.TrainTime
		// Expected outputs
		ExpectedCurrentTime time.Time
	}{
		"date_increments_when_crossing_midnight_forward": {
			StartDate:           time.Date(2025, 8, 10, 0, 0, 0, 0, location),   // 2025-08-10
			PreviousTime:        time.Date(2025, 8, 10, 23, 55, 0, 0, location), // 2025-08-10 23:55
			CurrentTime:         "00:05",
			ExpectedCurrentTime: time.Date(2025, 8, 11, 0, 5, 0, 0, location), // 2025-08-11 00:05
		},
		"time_can_go_forward": {
			StartDate:           time.Date(2025, 8, 10, 0, 0, 0, 0, location), // 2025-08-10
			PreviousTime:        time.Date(2025, 8, 10, 8, 0, 0, 0, location), // 2025-08-10 08:00
			CurrentTime:         "08:05",
			ExpectedCurrentTime: time.Date(2025, 8, 10, 8, 5, 0, 0, location), // 2025-08-10 08:05
		},
		"time_can_go_backward": {
			StartDate:           time.Date(2025, 8, 10, 0, 0, 0, 0, location), // 2025-08-10
			PreviousTime:        time.Date(2025, 8, 10, 8, 0, 0, 0, location), // 2025-08-10 08:00
			CurrentTime:         "07:55",
			ExpectedCurrentTime: time.Date(2025, 8, 10, 7, 55, 0, 0, location), // 2025-08-10 07:55
		},
		"date_decrements_when_crossing_midnight_backward": {
			StartDate:           time.Date(2025, 8, 10, 0, 0, 0, 0, location), // 2025-08-10
			PreviousTime:        time.Date(2025, 8, 11, 0, 5, 0, 0, location), // 2025-08-11 00:05
			CurrentTime:         "23:55",
			ExpectedCurrentTime: time.Date(2025, 8, 10, 23, 55, 0, 0, location), // 2025-08-10 23:55
		},
		"no_previoustime_uses_startdate": {
			StartDate:           time.Date(2025, 8, 10, 0, 0, 0, 0, location), // 2025-08-10
			PreviousTime:        time.Time{},                                  // zero
			CurrentTime:         "08:00",
			ExpectedCurrentTime: time.Date(2025, 8, 10, 8, 0, 0, 0, location), // 2025-08-10 08:00
		},
		"previoustime_used_for_date_when_given": {
			StartDate:           time.Date(2025, 8, 10, 0, 0, 0, 0, location), // 2025-08-10
			PreviousTime:        time.Date(2025, 8, 11, 8, 0, 0, 0, location), // 2025-08-11 08:00
			CurrentTime:         "08:05",
			ExpectedCurrentTime: time.Date(2025, 8, 11, 8, 5, 0, 0, location), // 2025-08-11 08:05
		},
		"time_can_contain_a_seconds_component": {
			StartDate:           time.Date(2025, 8, 10, 0, 0, 0, 0, location), // 2025-08-10
			PreviousTime:        time.Date(2025, 8, 10, 8, 0, 0, 0, location), // 2025-08-10 08:00
			CurrentTime:         "08:05:30",
			ExpectedCurrentTime: time.Date(2025, 8, 10, 8, 5, 30, 0, location), // 2025-08-10 08:05:30
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			time, err := trainTimeToTime(tc.PreviousTime, tc.CurrentTime, tc.StartDate)
			if err != nil {
				t.Errorf("failed to convert TrainTime to time.Time: %v", err)
			}
			if !time.Equal(tc.ExpectedCurrentTime) {
				t.Errorf("expected time %v, got %v", tc.ExpectedCurrentTime, time)
			}
		})
	}
}
