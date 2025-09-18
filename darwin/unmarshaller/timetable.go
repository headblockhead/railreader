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
		return fmt.Errorf("failed to decode ScheduleInformation: %w", err)
	}

	// Convert the alias back to the original type.
	*j = Journey(journey)

	return nil
}

// TimetableLocation is a generic struct that contains (nilable pointers to) all the possible location types.
type TimetableLocation struct {
	Type LocationType

	Origin                  *OriginLocation                  `xml:"OR"`
	OperationalOrigin       *OperationalOriginLocation       `xml:"OPOR"`
	Intermediate            *IntermediateLocation            `xml:"IP"`
	OperationalIntermediate *OperationalIntermediateLocation `xml:"OPIP"`
	IntermediatePassing     *IntermediatePassingLocation     `xml:"PP"`
	Destination             *DestinationLocation             `xml:"DT"`
	OperationalDestination  *OperationalDestinationLocation  `xml:"OPDT"`
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
	PublicArrivalTime *TrainTime `xml:"pta,attr"`
	// PublicDepartureTime is optionally provided.
	PublicDepartureTime *TrainTime `xml:"ptd,attr"`
	// WorkingArrivalTime is optionally provided.
	WorkingArrivalTime   *TrainTime `xml:"wta,attr"`
	WorkingDepartureTime TrainTime  `xml:"wtd,attr"`
	// FalseDestination is an optionally provided destination TIPLOC that is not the train's true destination, but should be displayed to the public as the train's destination, at this location.
	FalseDestination *railreader.TimingPointLocationCode `xml:"fd,attr"`
}

type OperationalOriginTimetableLocation struct {
	TimetableLocationBase
	// WorkingArrivalTime is optionally provided.
	WorkingArrivalTime   *TrainTime `xml:"wta,attr"`
	WorkingDepartureTime TrainTime  `xml:"wtd,attr"`
}

type IntermediateTimetableLocation struct {
	TimetableLocationBase
	// PublicArrivalTime is optionally provided.
	PublicArrivalTime *TrainTime `xml:"pta,attr"`
	// PublicDepartureTime is optionally provided.
	PublicDepartureTime  *TrainTime `xml:"ptd,attr"`
	WorkingArrivalTime   TrainTime  `xml:"wta,attr"`
	WorkingDepartureTime TrainTime  `xml:"wtd,attr"`
	// RoutingDelay is an optionally provided amount of minutes a change in the train's routing has delayed this location's PublicArrivalTime.
	RoutingDelay *int `xml:"rdelay,attr"`
	// FalseDestination is an optionally provided destination TIPLOC that is not the train's true destination, but should be displayed to the public as the train's destination, at this location.
	FalseDestination *railreader.TimingPointLocationCode `xml:"fd,attr"`
}

type OperationalIntermediateTimetableLocation struct {
	TimetableLocationBase
	WorkingArrivalTime   TrainTime `xml:"wta,attr"`
	WorkingDepartureTime TrainTime `xml:"wtd,attr"`
	// RoutingDelay is an optionally provided amount of minutes a change in the train's routing has delayed this location's PublicArrivalTime.
	RoutingDelay *int `xml:"rdelay,attr"`
}

type IntermediatePassingTimetableLocation struct {
	TimetableLocationBase
	WorkingPassingTime TrainTime `xml:"wtp,attr"`
	// RoutingDelay is an optionally provided amount of minutes a change in the train's routing has delayed this location's PublicArrivalTime.
	RoutingDelay *int `xml:"rdelay,attr"`
}

type DestinationTimetableLocation struct {
	TimetableLocationBase
	// PublicArrivalTime is optionally provided.
	PublicArrivalTime *TrainTime `xml:"pta,attr"`
	// PublicDepartureTime is optionally provided.
	PublicDepartureTime *TrainTime `xml:"ptd,attr"`
	WorkingArrivalTime  TrainTime  `xml:"wta,attr"`
	// WorkingDepartureTime is optionally provided.
	WorkingDepartureTime *TrainTime `xml:"wtd,attr"`
	// RoutingDelay is an optionally provided amount of minutes a change in the train's routing has delayed this location's PublicArrivalTime.
	RoutingDelay *int `xml:"rdelay,attr"`
}

type OperationalDestinationTimetableLocation struct {
	TimetableLocationBase
	WorkingArrivalTime TrainTime `xml:"wta,attr"`
	// WorkingDepartureTime is optionally provided.
	WorkingDepartureTime *TrainTime `xml:"wtd,attr"`
	// RoutingDelay is an optionally provided amount of minutes a change in the train's routing has delayed this location's PublicArrivalTime.
	RoutingDelay *int `xml:"rdelay,attr"`
}
