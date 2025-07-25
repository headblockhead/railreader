package decoder

import (
	"encoding/xml"

	"github.com/headblockhead/railreader"
)

// trueIfPresent implements xml.Unmarshaler.
// It unmarshals to true if the element of this type exists.
type trueIfPresent bool

func (p *trueIfPresent) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	if start.Name.Local != "" {
		*p = true
	}
	d.Skip()
	return nil
}

// LocationTimeIdentifiers is used as a base struct.
// It helps to identify a specific stop on a train's schedule.
// This can be done by matching the time(s) present with a specific TIPLOC to get a stop on a schedule.
// This is useful for circular routes, where a train may visit the same TIPLOC multiple times in a single schedule.
type LocationTimeIdentifiers struct {
	// at least one of:
	WorkingArrivalTime   railreader.TrainTime `xml:"wta,attr"`
	WorkingDepartureTime railreader.TrainTime `xml:"wtd,attr"`
	WorkingPassingTime   railreader.TrainTime `xml:"wtp,attr"`
	PublicArrivalTime    railreader.TrainTime `xml:"pta,attr"`
	PublicDepartureTime  railreader.TrainTime `xml:"ptd,attr"`
}

// TrainIdentifiers is used as a base struct.
// RID is enough to identify a specific train (as far as I know), but UID and ScheduledStartDate are included for additional context.
type TrainIdentifiers struct {
	// RID is the unique 16-character ID for a specific train.
	RID string `xml:"rid,attr"`
	// UID is a 6-character ID for a scheduled route.
	UID string `xml:"uid,attr"`
	// ScheduledStartDate is the date the train started in YYYY-MM-DD format.
	ScheduledStartDate string `xml:"ssd,attr"`
}

// CISCode is a code that identifies the ID of the system that sent the request.
// A mapping of CIS codes to system names is included in the reference data.
type CISCode string
