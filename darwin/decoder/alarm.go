package decoder

import (
	"github.com/headblockhead/railreader"
)

// Alarm is a report of an internal issue/failure of a datasource that feeds Darwin.
type Alarm struct {
	// only one of:
	// ClearedAlarm may contain the ID of an alarm that has been cleared.
	ClearedAlarm int       `xml:"clear"`
	NewAlarm     *NewAlarm `xml:"set"`
}

type NewAlarm struct {
	ID int `xml:"id,attr"`

	// only one of:
	TrainDescriberAreaFailure      railreader.TrainDescriberArea `xml:"tdAreaFail"`
	TrainDescriberTotalFeedFailure TrueIfPresent                 `xml:"tdFeedFail"`
	TyrellTotalFeedFailure         TrueIfPresent                 `xml:"tyrellFeedFail"`
}
