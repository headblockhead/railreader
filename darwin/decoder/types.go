package decoder

import (
	"encoding/xml"
	"fmt"
	"time"

	"github.com/headblockhead/railreader"
)

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
	WorkingArrivalTime   TrainTime `xml:"wta,attr"`
	WorkingDepartureTime TrainTime `xml:"wtd,attr"`
	WorkingPassingTime   TrainTime `xml:"wtp,attr"`
	PublicArrivalTime    TrainTime `xml:"pta,attr"`
	PublicDepartureTime  TrainTime `xml:"ptd,attr"`
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
	TIPLOC railreader.TimingPointLocationCode `xml:"tiploc,attr"`
	// Near is true if the disruption should be interpreted as being near the provided TIPLOC, rather than at the exact location.
	Near bool `xml:"near,attr"`

	ReasonID int `xml:",chardata"`
}

// CISCode (Customer Information System Code) is a code that identifies the ID of the system that sent the request.
// A mapping of CIS codes to system names is included in the reference data.
type CISCode string

// CRSCode (Computerised Reservation System Code) is a 3-letter code that identifies a passenger rail station.
type CRSCode string

// TrainOperatingCompanyCode is a two-letter code.
type TrainOperatingCompanyCode string

// TrainTime is formatted as HH:MM:SS or HH:MM.
type TrainTime string

func (t TrainTime) Time(previousTime TrainTime, date time.Time) (*time.Time, error) {
	if previousTime == "" {
		previousTime = t
	}

	location, err := time.LoadLocation("Europe/London")
	if err != nil {
		return nil, fmt.Errorf("failed to load location: %w", err)
	}

	parsedCurrentTime, err := time.ParseInLocation(time.TimeOnly, string(t), location)
	if err != nil {
		return nil, fmt.Errorf("failed to parse time %q: %w", t, err)
	}
	parsedPreviousTime, err := time.ParseInLocation(time.TimeOnly, string(previousTime), location)
	if err != nil {
		return nil, fmt.Errorf("failed to parse previous time %q: %w", previousTime, err)
	}

	difference := parsedCurrentTime.Sub(parsedPreviousTime)

	// Crossed midnight forwards
	if difference < -6*time.Hour {
		date = date.AddDate(0, 0, 1)
	}
	// Backwards time
	if difference >= -6*time.Hour && difference <= 0 {
	}
	// Normal time
	if difference >= 0 && difference <= 18*time.Hour {
	}
	// Crossed midnight backwards
	if difference > 18*time.Hour {
		date = date.AddDate(0, 0, -1)
	}

	finalTime := parsedCurrentTime.AddDate(date.Year(), int(date.Month()-1), date.Day()-1)

	return &finalTime, nil
}
