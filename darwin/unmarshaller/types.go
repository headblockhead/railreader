package unmarshaller

import (
	"encoding/xml"
)

func pointerTo[T any](v T) *T {
	return &v
}

// TrueIfPresent implements xml.Unmarshaler.
// It unmarshals to true if the element of this type exists.
type TrueIfPresent bool

func (p *TrueIfPresent) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	if start.Name.Local != "" {
		*p = true
	}
	if err := d.Skip(); err != nil {
		return err
	}
	return nil
}

// LocationTimeIdentifiers is used as a base struct.
// It helps to identify a specific stop on a train's schedule.
// This can be done by matching the time(s) present with a specific TIPLOC to get a stop on a schedule.
// This is useful for circular routes, where a train may visit the same TIPLOC multiple times in a single schedule.
type LocationTimeIdentifiers struct {
	// at least one of:
	WorkingArrivalTime   *string `xml:"wta,attr"`
	WorkingDepartureTime *string `xml:"wtd,attr"`
	WorkingPassingTime   *string `xml:"wtp,attr"`
	PublicArrivalTime    *string `xml:"pta,attr"`
	PublicDepartureTime  *string `xml:"ptd,attr"`
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

type DisruptionReason struct {
	// TIPLOC is the optionally provided location code for the position of the disruption.
	TIPLOC *string `xml:"tiploc,attr"`
	// Near is true if the disruption should be interpreted as being near the provided TIPLOC, rather than at the exact location.
	Near bool `xml:"near,attr"`

	ReasonID int `xml:",chardata"`
}
