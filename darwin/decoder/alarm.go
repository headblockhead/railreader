package decoder

import "github.com/headblockhead/railreader"

// Alarm is a report of an internal issue/failure of a datasource that feeds Darwin.
type Alarm struct {
	// ClearedAlarm is the ID of an alarm that has been cleared.
	ClearedAlarm string   `xml:"clear"`
	NewAlarm     NewAlarm `xml:"set"`
}

type NewAlarm struct {
	ID string `xml:"id,attr"`

	// contains one of:
	TrainDescriptorAreaFailure railreader.TrainDescriptorArea `xml:"tdAreaFail"`
	TrainDescriptorFeedFailure *string                        `xml:"tdFeedFail"`
	TyrellFeedFailure          *string                        `xml:"tyrellFeedFail"`
}
