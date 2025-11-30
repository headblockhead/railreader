package unmarshaller

import (
	"encoding/xml"
	"fmt"
	"io"

	"github.com/headblockhead/railreader"
)

const ExpectedTimetableFileSuffix = "_v8.xml.gz"

// Timetable version 8
type Timetable struct {
	ID           string        `xml:"timetableID,attr"`
	Journeys     []Journey     `xml:"Journey"`
	Associations []Association `xml:"Association"`
}

func NewTimetable(xmlData io.Reader) (*Timetable, error) {
	decoder := xml.NewDecoder(xmlData)
	var tt Timetable
	if err := decoder.Decode(&tt); err != nil {
		return nil, err
	}
	return &tt, nil
}

// Journey is very similar to Service, but without some fields, and with a few extras.
type Journey struct {
	TrainIdentifiers
	// Headcode is the 4-character headcode of the train, with the format:
	// [0-9][A-Z][0-9][0-9]
	Headcode string `xml:"trainId,attr"`
	// TOC is the Rail Delivery Group's 2-character code for the train operating company.
	TOC string `xml:"toc,attr"`
	// Service is the 1-character code for the type of transport.
	// If not provided, it defaults to a Passenger and Parcel Train.
	Service railreader.ServiceType `xml:"status,attr"`
	// Category is a 2-character code for the load of the service.
	// If not provided, it defaults to OO.
	Category railreader.ServiceCategory `xml:"trainCat,attr"`
	// PassengerService is true if not provided. This will sometimes be false, based on the Category.
	PassengerService bool `xml:"isPassengerSvc,attr"`
	// Deleted means you should not use or display this schedule.
	Deleted bool `xml:"deleted,attr"`
	Charter bool `xml:"isCharter,attr"`
	// QTrains are trains that only run as-required, and have not yet been activated.
	QTrain    bool `xml:"isQTrain,attr"`
	Cancelled bool `xml:"can,attr"`

	Locations []JourneyLocation `xml:",any"`
	// CancellationReason is the optionally provided reason why this service was cancelled.
	// This is provided at the service level, and/or the location level.
	CancellationReason *DisruptionReason `xml:"cancelReason"`
}

func (j *Journey) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	// Alias type created used to avoid recursion.
	type Alias Journey
	var journey Alias

	// Set default values.
	journey.Service = railreader.ServicePassengerOrParcelTrain
	journey.Category = railreader.CategoryPassenger
	journey.PassengerService = true

	if err := d.DecodeElement(&journey, &start); err != nil {
		return fmt.Errorf("failed to decode Journey: %w", err)
	}

	// Convert the alias back to the original type.
	*j = Journey(journey)

	return nil
}

// JourneyLocation is a generic struct that contains (nilable pointers to) all the possible location types.
type JourneyLocation struct {
	Type LocationType

	Origin                  *JourneyOriginLocation                  `xml:"OR"`
	OperationalOrigin       *JourneyOperationalOriginLocation       `xml:"OPOR"`
	Intermediate            *JourneyIntermediateLocation            `xml:"IP"`
	OperationalIntermediate *JourneyOperationalIntermediateLocation `xml:"OPIP"`
	IntermediatePassing     *JourneyIntermediatePassingLocation     `xml:"PP"`
	Destination             *JourneyDestinationLocation             `xml:"DT"`
	OperationalDestination  *JourneyOperationalDestinationLocation  `xml:"OPDT"`
}

func (lg *JourneyLocation) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	locationType := start.Name.Local
	lg.Type = LocationType(locationType)
	switch lg.Type {
	case LocationTypeOrigin:
		lg.Origin = &JourneyOriginLocation{JourneyLocationBase: JourneyLocationBase{}}
		if err := d.DecodeElement(lg.Origin, &start); err != nil {
			return fmt.Errorf("failed to decode JourneyOriginLocation: %w", err)
		}
	case LocationTypeOperationalOrigin:
		lg.OperationalOrigin = &JourneyOperationalOriginLocation{JourneyLocationBase: JourneyLocationBase{}}
		if err := d.DecodeElement(lg.OperationalOrigin, &start); err != nil {
			return fmt.Errorf("failed to decode JourneyOperationalOriginLocation: %w", err)
		}
	case LocationTypeIntermediate:
		lg.Intermediate = &JourneyIntermediateLocation{JourneyLocationBase: JourneyLocationBase{}}
		if err := d.DecodeElement(lg.Intermediate, &start); err != nil {
			return fmt.Errorf("failed to decode JourneyIntermediateLocation: %w", err)
		}
	case LocationTypeOperationalIntermediate:
		lg.OperationalIntermediate = &JourneyOperationalIntermediateLocation{JourneyLocationBase: JourneyLocationBase{}}
		if err := d.DecodeElement(lg.OperationalIntermediate, &start); err != nil {
			return fmt.Errorf("failed to decode JourneyOperationalIntermediateLocation: %w", err)
		}
	case LocationTypeIntermediatePassing:
		lg.IntermediatePassing = &JourneyIntermediatePassingLocation{JourneyLocationBase: JourneyLocationBase{}}
		if err := d.DecodeElement(lg.IntermediatePassing, &start); err != nil {
			return fmt.Errorf("failed to decode JourneyIntermediatePassingLocation: %w", err)
		}
	case LocationTypeDestination:
		lg.Destination = &JourneyDestinationLocation{JourneyLocationBase: JourneyLocationBase{}}
		if err := d.DecodeElement(lg.Destination, &start); err != nil {
			return fmt.Errorf("failed to decode JourneyDestinationLocation: %w", err)
		}
	case LocationTypeOperationalDestination:
		lg.OperationalDestination = &JourneyOperationalDestinationLocation{JourneyLocationBase: JourneyLocationBase{}}
		if err := d.DecodeElement(lg.OperationalDestination, &start); err != nil {
			return fmt.Errorf("failed to decode JourneyOperationalDestinationLocation: %w", err)
		}
	default:
		return fmt.Errorf("unknown location type: %s", locationType)
	}

	return nil
}

type JourneyLocationBase struct {
	// TIPLOC is the code for the location
	TIPLOC string `xml:"tpl,attr"`
	// Activities optionally provides what is happening at this location.
	// Activities can be converted into a slice of railreader.ActivityCode.
	// If it is empty, it should be interpreted as a slice containing 1 railreader.ActivityNone.
	Activities *string `xml:"act,attr"`
	// PlannedActivities optionally provides what was/is planned to happen at this location.
	// This is only usually given if the Activity is different to the PlannedActivities.
	// PlannedActivities can be converted into a slice of railreader.ActivityCode.
	// If it is empty, it should be interpreted as a slice containing 1 railreader.ActivityNone.
	PlannedActivities *string `xml:"planAct,attr"`
	Cancelled         bool    `xml:"can,attr"`
	Platform          *string `xml:"plat,attr"`
}

type JourneyOriginLocation struct {
	OriginLocation
	JourneyLocationBase
}

type JourneyOperationalOriginLocation struct {
	OperationalOriginLocation
	JourneyLocationBase
}

type JourneyIntermediateLocation struct {
	IntermediateLocation
	JourneyLocationBase
}

type JourneyOperationalIntermediateLocation struct {
	OperationalIntermediateLocation
	JourneyLocationBase
}

type JourneyIntermediatePassingLocation struct {
	IntermediatePassingLocation
	JourneyLocationBase
}

type JourneyDestinationLocation struct {
	DestinationLocation
	JourneyLocationBase
}

type JourneyOperationalDestinationLocation struct {
	OperationalDestinationLocation
	JourneyLocationBase
}
