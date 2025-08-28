package unmarshaller

import (
	"github.com/headblockhead/railreader"
)

// Alarm is a report of an internal issue/failure of a data-source that feeds Darwin.
type Alarm struct {
	// only one of:
	// ClearedAlarm may contain the ID of an alarm that has been cleared.
	ClearedAlarm int       `xml:"clear"`
	NewAlarm     *NewAlarm `xml:"set"`
}

type NewAlarm struct {
	ID int `xml:"id,attr"`

	// only one of:
	// TDFailure gives the code of a specific Train Describer that Darwin has not received any data from for a period of time (darwin suspects that the TD has failed)
	TDFailure railreader.TrainDescriber `xml:"tdAreaFail"`
	// TDTotalFailure is true if Darwin has not received any data from any Train Describers for a period of time.
	TDTotalFailure TrueIfPresent `xml:"tdFeedFail"`
	// TyrellTotalFailure is true if Darwin's connection to Tyrell (a disruption alert notification system - see https://www.nexusalpha.com/tyrell-io) has failed.
	TyrellTotalFailure TrueIfPresent `xml:"tyrellFeedFail"`
}
