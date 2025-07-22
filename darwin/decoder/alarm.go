package decoder

import (
	"encoding/xml"

	"github.com/headblockhead/railreader"
)

// Alarm is a report of an internal issue/failure of a datasource that feeds Darwin.
type Alarm struct {
	// contains only one of:
	// ClearedAlarm may contain the ID of an alarm that has been cleared.
	ClearedAlarm string    `xml:"clear"`
	NewAlarm     *NewAlarm `xml:"set"`
}

type NewAlarm struct {
	ID string `xml:"id,attr"`

	// contains only one of:
	TrainDescriptorAreaFailure railreader.TrainDescriptorArea `xml:"tdAreaFail"`
	TrainDescriptorFeedFailure trueIfPresent                  `xml:"tdFeedFail"`
	TyrellFeedFailure          trueIfPresent                  `xml:"tyrellFeedFail"`
}

type trueIfPresent bool

func (p *trueIfPresent) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	if start.Name.Local != "" {
		*p = true
	}
	d.Skip()
	return nil
}
