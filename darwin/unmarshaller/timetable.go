package unmarshaller

import (
	"encoding/xml"
	"fmt"

	"github.com/headblockhead/railreader"
)

// Timetable version 8
type Timetable struct {
	ID           string        `xml:"timetableID,attr"`
	Journeys     []Journey     `xml:"Journey"`
	Associations []Association `xml:"Association"`
}

func NewTimetable(xmlData string) (tt Timetable, err error) {
	if err = xml.Unmarshal([]byte(xmlData), &tt); err != nil {
		return
	}
	return
}

// Journey is very similar to Service, but is missing some fields, and has a few extras.
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

	Locations []TimetableLocation `xml:",any"`
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

// TimetableLocation is a generic struct that contains (nilable pointers to) all the possible location types.
type TimetableLocation struct {
	Type LocationType

	Origin                  *OriginTimetableLocation                  `xml:"OR"`
	OperationalOrigin       *OperationalOriginTimetableLocation       `xml:"OPOR"`
	Intermediate            *IntermediateTimetableLocation            `xml:"IP"`
	OperationalIntermediate *OperationalIntermediateTimetableLocation `xml:"OPIP"`
	IntermediatePassing     *IntermediatePassingTimetableLocation     `xml:"PP"`
	Destination             *DestinationTimetableLocation             `xml:"DT"`
	OperationalDestination  *OperationalDestinationTimetableLocation  `xml:"OPDT"`
}

func (lg *TimetableLocation) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	locationType := start.Name.Local
	lg.Type = LocationType(locationType)
	switch lg.Type {
	case LocationTypeOrigin:
		lg.Origin = &OriginTimetableLocation{TimetableLocationBase: TimetableLocationBase{}}
		if err := d.DecodeElement(lg.Origin, &start); err != nil {
			return fmt.Errorf("failed to decode OriginTimetableLocation: %w", err)
		}
	case LocationTypeOperationalOrigin:
		lg.OperationalOrigin = &OperationalOriginTimetableLocation{TimetableLocationBase: TimetableLocationBase{}}
		if err := d.DecodeElement(lg.OperationalOrigin, &start); err != nil {
			return fmt.Errorf("failed to decode OperationalOriginTimetableLocation: %w", err)
		}
	case LocationTypeIntermediate:
		lg.Intermediate = &IntermediateTimetableLocation{TimetableLocationBase: TimetableLocationBase{}}
		if err := d.DecodeElement(lg.Intermediate, &start); err != nil {
			return fmt.Errorf("failed to decode IntermediateTimetableLocation: %w", err)
		}
	case LocationTypeOperationalIntermediate:
		lg.OperationalIntermediate = &OperationalIntermediateTimetableLocation{TimetableLocationBase: TimetableLocationBase{}}
		if err := d.DecodeElement(lg.OperationalIntermediate, &start); err != nil {
			return fmt.Errorf("failed to decode OperationalIntermediateTimetableLocation: %w", err)
		}
	case LocationTypeIntermediatePassing:
		lg.IntermediatePassing = &IntermediatePassingTimetableLocation{TimetableLocationBase: TimetableLocationBase{}}
		if err := d.DecodeElement(lg.IntermediatePassing, &start); err != nil {
			return fmt.Errorf("failed to decode IntermediatePassingTimetableLocation: %w", err)
		}
	case LocationTypeDestination:
		lg.Destination = &DestinationTimetableLocation{TimetableLocationBase: TimetableLocationBase{}}
		if err := d.DecodeElement(lg.Destination, &start); err != nil {
			return fmt.Errorf("failed to decode DestinationTimetableLocation: %w", err)
		}
	case LocationTypeOperationalDestination:
		lg.OperationalDestination = &OperationalDestinationTimetableLocation{TimetableLocationBase: TimetableLocationBase{}}
		if err := d.DecodeElement(lg.OperationalDestination, &start); err != nil {
			return fmt.Errorf("failed to decode OperationalDestinationTimetableLocation: %w", err)
		}
	default:
		return fmt.Errorf("unknown location type: %s", locationType)
	}

	return nil
}

type TimetableLocationBase struct {
	// TIPLOC is the code for the location
	TIPLOC railreader.TimingPointLocationCode `xml:"tpl,attr"`
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

type OriginTimetableLocation struct {
	TimetableLocationBase
	// PublicArrivalTime is optionally provided.
	PublicArrivalTime *string `xml:"pta,attr"`
	// PublicDepartureTime is optionally provided.
	PublicDepartureTime *string `xml:"ptd,attr"`
	// WorkingArrivalTime is optionally provided.
	WorkingArrivalTime   *string `xml:"wta,attr"`
	WorkingDepartureTime string  `xml:"wtd,attr"`
	// FalseDestination is an optionally provided destination TIPLOC that is not the train's true destination, but should be displayed to the public as the train's destination, at this location.
	FalseDestination *railreader.TimingPointLocationCode `xml:"fd,attr"`
}

type OperationalOriginTimetableLocation struct {
	TimetableLocationBase
	// WorkingArrivalTime is optionally provided.
	WorkingArrivalTime   *string `xml:"wta,attr"`
	WorkingDepartureTime string  `xml:"wtd,attr"`
}

type IntermediateTimetableLocation struct {
	TimetableLocationBase
	// PublicArrivalTime is optionally provided.
	PublicArrivalTime *string `xml:"pta,attr"`
	// PublicDepartureTime is optionally provided.
	PublicDepartureTime  *string `xml:"ptd,attr"`
	WorkingArrivalTime   string  `xml:"wta,attr"`
	WorkingDepartureTime string  `xml:"wtd,attr"`
	// RoutingDelay is an optionally provided amount of minutes a change in the train's routing has delayed this location's PublicArrivalTime.
	RoutingDelay *int `xml:"rdelay,attr"`
	// FalseDestination is an optionally provided destination TIPLOC that is not the train's true destination, but should be displayed to the public as the train's destination, at this location.
	FalseDestination *railreader.TimingPointLocationCode `xml:"fd,attr"`
}

type OperationalIntermediateTimetableLocation struct {
	TimetableLocationBase
	WorkingArrivalTime   string `xml:"wta,attr"`
	WorkingDepartureTime string `xml:"wtd,attr"`
	// RoutingDelay is an optionally provided amount of minutes a change in the train's routing has delayed this location's PublicArrivalTime.
	RoutingDelay *int `xml:"rdelay,attr"`
}

type IntermediatePassingTimetableLocation struct {
	TimetableLocationBase
	WorkingPassingTime string `xml:"wtp,attr"`
	// RoutingDelay is an optionally provided amount of minutes a change in the train's routing has delayed this location's PublicArrivalTime.
	RoutingDelay *int `xml:"rdelay,attr"`
}

type DestinationTimetableLocation struct {
	TimetableLocationBase
	// PublicArrivalTime is optionally provided.
	PublicArrivalTime *string `xml:"pta,attr"`
	// PublicDepartureTime is optionally provided.
	PublicDepartureTime *string `xml:"ptd,attr"`
	WorkingArrivalTime  string  `xml:"wta,attr"`
	// WorkingDepartureTime is optionally provided.
	WorkingDepartureTime *string `xml:"wtd,attr"`
	// RoutingDelay is an optionally provided amount of minutes a change in the train's routing has delayed this location's PublicArrivalTime.
	RoutingDelay *int `xml:"rdelay,attr"`
}

type OperationalDestinationTimetableLocation struct {
	TimetableLocationBase
	WorkingArrivalTime string `xml:"wta,attr"`
	// WorkingDepartureTime is optionally provided.
	WorkingDepartureTime *string `xml:"wtd,attr"`
	// RoutingDelay is an optionally provided amount of minutes a change in the train's routing has delayed this location's PublicArrivalTime.
	RoutingDelay *int `xml:"rdelay,attr"`
}
