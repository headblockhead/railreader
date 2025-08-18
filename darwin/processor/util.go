package processor

import (
	"fmt"
	"time"

	"github.com/headblockhead/railreader/darwin/decoder"
)

func trainTimeToTime(previousTime time.Time, currentTrainTime decoder.TrainTime, scheduledStartDate time.Time) (*time.Time, error) {
	if len(currentTrainTime) != 5 && len(currentTrainTime) != 8 {
		return nil, fmt.Errorf("invalid train time length of: %q", currentTrainTime)
	}

	template := "15:04"
	if len(currentTrainTime) == 8 {
		template = "15:04:05"
	}

	location, err := time.LoadLocation("Europe/London")
	if err != nil {
		return nil, fmt.Errorf("failed to load time location: %w", err)
	}
	currentTime, err := time.ParseInLocation(template, string(currentTrainTime), location)
	if err != nil {
		return nil, fmt.Errorf("failed to parse time %q: %w", currentTrainTime, err)
	}

	currentTime = time.Date(scheduledStartDate.Year(), scheduledStartDate.Month(), scheduledStartDate.Day(), currentTime.Hour(), currentTime.Minute(), currentTime.Second(), 0, location)

	if previousTime.IsZero() {
		return &currentTime, nil
	}

	difference := currentTime.Sub(previousTime)

	// Crossed midnight forwards
	if difference < -6*time.Hour {
		scheduledStartDate = scheduledStartDate.AddDate(0, 0, 1)
	}
	// Backwards time
	if difference < 0 && difference >= -6*time.Hour {
	}
	// Normal time
	if difference >= 0 && difference <= 18*time.Hour {
	}
	// Crossed midnight backwards
	if difference > 18*time.Hour {
		scheduledStartDate = scheduledStartDate.AddDate(0, 0, -1)
	}

	finalTime := time.Date(scheduledStartDate.Year(), scheduledStartDate.Month(), scheduledStartDate.Day(), currentTime.Hour(), currentTime.Minute(), currentTime.Second(), 0, location)

	return &finalTime, nil
}
