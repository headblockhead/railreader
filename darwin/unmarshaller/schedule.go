package unmarshaller

import (
	"encoding/xml"
	"fmt"

	"github.com/headblockhead/railreader"
)

type Schedule struct {
	TrainIdentifiers
	// Headcode is the 4-character headcode of the train, with the format:
	// [0-9][A-Z][0-9][0-9]
	Headcode string `xml:"trainId,attr"`
	// RetailServiceID is the optionally provided Retail Service ID, either as an:
	// 8 character "portion identifier" (including a leading TOC code),
	// or a 6 character "base identifier" (without a TOC code).
	RetailServiceID *string `xml:"rsid,attr"`
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
	// Active is true if not provided. Active is sometimes set to false when a service has been deactivated by a DeactivationInformation element, but Active is only set in snapshots.
	Active bool `xml:"isActive,attr"`
	// Deleted means you should not use or display this schedule.
	Deleted bool `xml:"deleted,attr"`
	Charter bool `xml:"isCharter,attr"`

	// Locations is a slice of at least 2 location elements that describe the train's schedule.
	Locations []ScheduleLocation `xml:",any"` // Any other provided XML elements will be interpreted as locations.
	// CancellationReason is the optionally provided reason why this service was cancelled.
	// This is provided at the service level, and/or the location level.
	CancellationReason *DisruptionReason `xml:"cancelReason"`
	// DivertedVia is the optionally provided TIPLOC that this service has been diverted via (which may or may not be on the timetable).
	DivertedVia *railreader.TimingPointLocationCode `xml:"divertedVia"`
	// DiversionReason is the optionally provided reason why this service has been diverted.
	DiversionReason *DisruptionReason `xml:"diversionReason"`
}

func (si *Schedule) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	// Alias type created used to avoid recursion.
	type Alias Schedule
	var schedule Alias

	// Set default values.
	schedule.Service = railreader.ServicePassengerOrParcelTrain
	schedule.Category = railreader.CategoryPassenger
	schedule.PassengerService = true
	schedule.Active = true

	if err := d.DecodeElement(&schedule, &start); err != nil {
		return fmt.Errorf("failed to decode Schedule: %w", err)
	}

	// Convert the alias back to the original type.
	*si = Schedule(schedule)

	return nil
}

// ScheduleLocation is a generic struct that contains (nilable pointers to) all the possible location types.
type ScheduleLocation struct {
	Type LocationType

	Origin                  *OriginLocation                  `xml:"OR"`
	OperationalOrigin       *OperationalOriginLocation       `xml:"OPOR"`
	Intermediate            *IntermediateLocation            `xml:"IP"`
	OperationalIntermediate *OperationalIntermediateLocation `xml:"OPIP"`
	IntermediatePassing     *IntermediatePassingLocation     `xml:"PP"`
	Destination             *DestinationLocation             `xml:"DT"`
	OperationalDestination  *OperationalDestinationLocation  `xml:"OPDT"`
}

type LocationType string

const (
	LocationTypeOrigin                  LocationType = "OR"
	LocationTypeOperationalOrigin       LocationType = "OPOR"
	LocationTypeIntermediate            LocationType = "IP"
	LocationTypeOperationalIntermediate LocationType = "OPIP"
	LocationTypeIntermediatePassing     LocationType = "PP"
	LocationTypeDestination             LocationType = "DT"
	LocationTypeOperationalDestination  LocationType = "OPDT"
)

func (lg *ScheduleLocation) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	locationType := start.Name.Local
	lg.Type = LocationType(locationType)
	switch lg.Type {
	case LocationTypeOrigin:
		lg.Origin = &OriginLocation{LocationBase: LocationBase{}}
		if err := d.DecodeElement(lg.Origin, &start); err != nil {
			return fmt.Errorf("failed to decode OriginLocation: %w", err)
		}
	case LocationTypeOperationalOrigin:
		lg.OperationalOrigin = &OperationalOriginLocation{LocationBase: LocationBase{}}
		if err := d.DecodeElement(lg.OperationalOrigin, &start); err != nil {
			return fmt.Errorf("failed to decode OperationalOriginLocation: %w", err)
		}
	case LocationTypeIntermediate:
		lg.Intermediate = &IntermediateLocation{LocationBase: LocationBase{}}
		if err := d.DecodeElement(lg.Intermediate, &start); err != nil {
			return fmt.Errorf("failed to decode IntermediateLocation: %w", err)
		}
	case LocationTypeOperationalIntermediate:
		lg.OperationalIntermediate = &OperationalIntermediateLocation{LocationBase: LocationBase{}}
		if err := d.DecodeElement(lg.OperationalIntermediate, &start); err != nil {
			return fmt.Errorf("failed to decode OperationalIntermediateLocation: %w", err)
		}
	case LocationTypeIntermediatePassing:
		lg.IntermediatePassing = &IntermediatePassingLocation{LocationBase: LocationBase{}}
		if err := d.DecodeElement(lg.IntermediatePassing, &start); err != nil {
			return fmt.Errorf("failed to decode IntermediatePassingLocation: %w", err)
		}
	case LocationTypeDestination:
		lg.Destination = &DestinationLocation{LocationBase: LocationBase{}}
		if err := d.DecodeElement(lg.Destination, &start); err != nil {
			return fmt.Errorf("failed to decode DestinationLocation: %w", err)
		}
	case LocationTypeOperationalDestination:
		lg.OperationalDestination = &OperationalDestinationLocation{LocationBase: LocationBase{}}
		if err := d.DecodeElement(lg.OperationalDestination, &start); err != nil {
			return fmt.Errorf("failed to decode OperationalDestinationLocation: %w", err)
		}
	default:
		return fmt.Errorf("unknown location type: %s", locationType)
	}

	return nil
}

// LocationBase is the base struct for all location types.
type LocationBase struct {
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
	// FormationID is the optionally provided ID of the train formation that is used at this location.
	// Formations 'ripple' forward from locations with a FormationID, until the next cancelled location, or the next FormationID.
	FormationID         *string `xml:"fid,attr"`
	AffectedByDiversion bool    `xml:"affectedByDiversion,attr"`

	// CancellationReason is an optionally provided reason why this location was cancelled.
	CancellationReason *DisruptionReason `xml:"cancelReason"`
}

type OriginLocation struct {
	LocationBase
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

type OperationalOriginLocation struct {
	LocationBase
	// WorkingArrivalTime is optionally provided.
	WorkingArrivalTime   *TrainTime `xml:"wta,attr"`
	WorkingDepartureTime TrainTime  `xml:"wtd,attr"`
}

type IntermediateLocation struct {
	LocationBase
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

type OperationalIntermediateLocation struct {
	LocationBase
	WorkingArrivalTime   TrainTime `xml:"wta,attr"`
	WorkingDepartureTime TrainTime `xml:"wtd,attr"`
	// RoutingDelay is an optionally provided amount of minutes a change in the train's routing has delayed this location's PublicArrivalTime.
	RoutingDelay *int `xml:"rdelay,attr"`
}

type IntermediatePassingLocation struct {
	LocationBase
	WorkingPassingTime TrainTime `xml:"wtp,attr"`
	// RoutingDelay is an optionally provided amount of minutes a change in the train's routing has delayed this location's PublicArrivalTime.
	RoutingDelay *int `xml:"rdelay,attr"`
}

type DestinationLocation struct {
	LocationBase
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

type OperationalDestinationLocation struct {
	LocationBase
	WorkingArrivalTime TrainTime `xml:"wta,attr"`
	// WorkingDepartureTime is optionally provided.
	WorkingDepartureTime *TrainTime `xml:"wtd,attr"`
	// RoutingDelay is an optionally provided amount of minutes a change in the train's routing has delayed this location's PublicArrivalTime.
	RoutingDelay *int `xml:"rdelay,attr"`
}
