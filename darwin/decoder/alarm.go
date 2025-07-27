package decoder

import (
	"github.com/headblockhead/railreader"
)

// Alarm is a report of an internal issue/failure of a datasource that feeds Darwin.
type Alarm struct {
	// only one of:
	// ClearedAlarm may contain the ID of an alarm that has been cleared.
	ClearedAlarm string    `xml:"clear"`
	NewAlarm     *NewAlarm `xml:"set"`
}

type NewAlarm struct {
	ID string `xml:"id,attr"`

	// only one of:
	TrainDescriptorAreaFailure      railreader.TrainDescriberArea `xml:"tdAreaFail"`
	TrainDescriptorTotalFeedFailure TrueIfPresent                 `xml:"tdFeedFail"`
	TyrellTotalFeedFailure          TrueIfPresent                 `xml:"tyrellFeedFail"`
}
