package processor

import (
	"fmt"
	"time"

	"github.com/headblockhead/railreader/darwin/decoder"
)

func trainTimeToTime(previousTime time.Time, currentTrainTime decoder.TrainTime, date time.Time) (*time.Time, error) {
	location, err := time.LoadLocation("Europe/London")
	if err != nil {
		return nil, fmt.Errorf("failed to load time location: %w", err)
	}

	var currentTime time.Time
	if len(currentTrainTime) == 8 {
		currentTime, err = time.ParseInLocation("15:04:05", string(currentTrainTime), location)
		if err != nil {
			return nil, fmt.Errorf("failed to parse time %q: %w", currentTrainTime, err)
		}
	} else if len(currentTrainTime) == 5 {
		currentTime, err = time.ParseInLocation("15:04", string(currentTrainTime), location)
		if err != nil {
			return nil, fmt.Errorf("failed to parse time %q: %w", currentTrainTime, err)
		}
	} else {
		return nil, fmt.Errorf("invalid train time length %q", currentTrainTime)
	}
	currentTime = time.Date(date.Year(), date.Month(), date.Day(), currentTime.Hour(), currentTime.Minute(), currentTime.Second(), 0, location)

	if previousTime.IsZero() {
		previousTime = currentTime
	}

	difference := currentTime.Sub(previousTime)

	// Crossed midnight forwards
	if difference < -6*time.Hour {
		date = date.AddDate(0, 0, 1)
	}
	// Backwards time
	if difference < 0 && difference >= -6*time.Hour {
	}
	// Normal time
	if difference >= 0 && difference <= 18*time.Hour {
	}
	// Crossed midnight backwards
	if difference > 18*time.Hour {
		date = date.AddDate(0, 0, -1)
	}

	finalTime := time.Date(date.Year(), date.Month(), date.Day(), currentTime.Hour(), currentTime.Minute(), currentTime.Second(), 0, location)

	return &finalTime, nil
}
