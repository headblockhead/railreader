package decoder

import (
	"encoding/xml"
	"fmt"

	"github.com/headblockhead/railreader"
)

type Schedule struct {
	// RID is the unique 16-character ID for a specific train, running this schedule, at this time.
	RID string `xml:"rid,attr"`
	// UID is (despite the name) a non-unique 6-character ID for this route at this time of day.
	UID string `xml:"uid,attr"`
	// TrainID is the 4-character headcode of the train, with the format:
	// [0-9][A-Z][0-9][0-9]
	TrainID string `xml:"trainId,attr"`
	// RetailServiceID is the optionally provided Retail Service ID, either as an:
	// 8 character "portion identifier" (including a leading TOC code),
	// or a 6 character "base identifier" (without a TOC code).
	RetailServiceID string `xml:"rsid,attr"`
	// ScheduledStartDate in YYYY-MM-DD format.
	ScheduledStartDate string `xml:"ssd,attr"`
	// TrainOperatingCompany is the Rail Delivery Group's 2-character code for the train operating company.
	TrainOperatingCompany string `xml:"toc,attr"`
	// Service is the 1-character code for the type of transport.
	// If not provided, it defaults to P (Passenger and Parcel Train).
	Service railreader.ServiceType `xml:"status,attr"`
	// Category is a 2-character code for the load of the service.
	// If not provided, it defaults to OO.
	Category railreader.ServiceCategory `xml:"trainCat,attr"`
	// PassengerService is true if not provided. This will sometimes be false, based on the value of the TrainCategory.
	PassengerService bool `xml:"isPassengerSvc,attr"`
	// Active is true if not provided. It is only present in snapshots, used to indicate a service has been deactivated by a DeactivationInformation element.
	Active bool `xml:"isActive,attr"`
	// Deleted means you should not use or display this schedule.
	Deleted bool `xml:"deleted,attr"`
	Charter bool `xml:"isCharter,attr"`

	// Locations is a slice of at least 2 location elements that describe the train's schedule.
	Locations []LocationGeneric `xml:",any"` // Any other provided XML elements will be interpreted as locations.
	// CancellationReason is the reason why this service was cancelled.
	// This is provided at the service level, and/or the location level.
	CancellationReason *DisruptionReason `xml:"cancelReason"`
	// DivertedVia is the TIPLOC that this service has been diverted via (which may or may not be on the timetable).
	DivertedVia railreader.TIPLOC `xml:"divertedVia"`
	// DiversionReason is the reason why this service has been diverted.
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
		return fmt.Errorf("failed to decode ScheduleInformation: %w", err)
	}

	// Convert the alias back to the original type.
	*si = Schedule(schedule)

	return nil
}

type DisruptionReason struct {
	// TIPLOC is the optionally provided location code for the position of the disruption.
	TIPLOC railreader.TIPLOC `xml:"tiploc,attr"`
	// Near is true if the disruption should be interpreted as being near the provided TIPLOC, rather than at the exact location.
	Near bool `xml:"near,attr"`

	Reason railreader.DisruptionReasonID `xml:",chardata"`
}

// LocationGeneric is a generic struct that contains (nullable pointers to) all the possible location types.
type LocationGeneric struct {
	Type                            LocationType
	OriginLocation                  *OriginLocation                  `xml:"OR"`
	OperationalOriginLocation       *OperationalOriginLocation       `xml:"OPOR"`
	IntermediateLocation            *IntermediateLocation            `xml:"IP"`
	OperationalIntermediateLocation *OperationalIntermediateLocation `xml:"OPIP"`
	IntermediatePassingLocation     *IntermediatePassingLocation     `xml:"PP"`
	DestinationLocation             *DestinationLocation             `xml:"DT"`
	OperationalDestinationLocation  *OperationalDestinationLocation  `xml:"OPDT"`
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

func (lg *LocationGeneric) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	locationType := start.Name.Local
	lg.Type = LocationType(locationType)
	switch lg.Type {
	case LocationTypeOrigin:
		lg.OriginLocation = &OriginLocation{LocationSchedule: LocationSchedule{}}
		if err := d.DecodeElement(lg.OriginLocation, &start); err != nil {
			return fmt.Errorf("failed to decode OriginLocation: %w", err)
		}
	case LocationTypeOperationalOrigin:
		lg.OperationalOriginLocation = &OperationalOriginLocation{LocationSchedule: LocationSchedule{}}
		if err := d.DecodeElement(lg.OperationalOriginLocation, &start); err != nil {
			return fmt.Errorf("failed to decode OperationalOriginLocation: %w", err)
		}
	case LocationTypeIntermediate:
		lg.IntermediateLocation = &IntermediateLocation{LocationSchedule: LocationSchedule{}}
		if err := d.DecodeElement(lg.IntermediateLocation, &start); err != nil {
			return fmt.Errorf("failed to decode IntermediateLocation: %w", err)
		}
	case LocationTypeOperationalIntermediate:
		lg.OperationalIntermediateLocation = &OperationalIntermediateLocation{LocationSchedule: LocationSchedule{}}
		if err := d.DecodeElement(lg.OperationalIntermediateLocation, &start); err != nil {
			return fmt.Errorf("failed to decode OperationalIntermediateLocation: %w", err)
		}
	case LocationTypeIntermediatePassing:
		lg.IntermediatePassingLocation = &IntermediatePassingLocation{LocationSchedule: LocationSchedule{}}
		if err := d.DecodeElement(lg.IntermediatePassingLocation, &start); err != nil {
			return fmt.Errorf("failed to decode IntermediatePassingLocation: %w", err)
		}
	case LocationTypeDestination:
		lg.DestinationLocation = &DestinationLocation{LocationSchedule: LocationSchedule{}}
		if err := d.DecodeElement(lg.DestinationLocation, &start); err != nil {
			return fmt.Errorf("failed to decode DestinationLocation: %w", err)
		}
	case LocationTypeOperationalDestination:
		lg.OperationalDestinationLocation = &OperationalDestinationLocation{LocationSchedule: LocationSchedule{}}
		if err := d.DecodeElement(lg.OperationalDestinationLocation, &start); err != nil {
			return fmt.Errorf("failed to decode OperationalDestinationLocation: %w", err)
		}
	default:
		return fmt.Errorf("unknown location type: %s", locationType)
	}

	return nil
}

func (lg LocationGeneric) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	switch lg.Type {
	case LocationTypeOrigin:
		if lg.OriginLocation != nil {
			return e.EncodeElement(lg.OriginLocation, start)
		}
	case LocationTypeOperationalOrigin:
		if lg.OperationalOriginLocation != nil {
			return e.EncodeElement(lg.OperationalOriginLocation, start)
		}
	case LocationTypeIntermediate:
		if lg.IntermediateLocation != nil {
			return e.EncodeElement(lg.IntermediateLocation, start)
		}
	case LocationTypeOperationalIntermediate:
		if lg.OperationalIntermediateLocation != nil {
			return e.EncodeElement(lg.OperationalIntermediateLocation, start)
		}
	case LocationTypeIntermediatePassing:
		if lg.IntermediatePassingLocation != nil {
			return e.EncodeElement(lg.IntermediatePassingLocation, start)
		}
	case LocationTypeDestination:
		if lg.DestinationLocation != nil {
			return e.EncodeElement(lg.DestinationLocation, start)
		}
	case LocationTypeOperationalDestination:
		if lg.OperationalDestinationLocation != nil {
			return e.EncodeElement(lg.OperationalDestinationLocation, start)
		}
	default:
		return fmt.Errorf("unknown location type: %s", lg.Type)
	}

	return nil
}

// LocationSchedule is the base struct for all location types.
type LocationSchedule struct {
	// TIPLOC is the code for the location
	TIPLOC railreader.TIPLOC `xml:"tpl,attr"`
	// Activities optionally provides what is happening at this location. It can be converted into a slice of railreader.ActivityCode.
	Activities string `xml:"act,attr"`
	// PlannedActivities optionally provides what was/is planned to happen at this location.
	// This is only usually given if the Activity is different to the PlannedActivities.
	// It is can be converted into a slice of railreader.ActivityCode.
	PlannedActivities string `xml:"planAct,attr"`
	Cancelled         bool   `xml:"can,attr"`
	// FormationID is the ID of the train formation that is used at this location.
	// Formations 'ripple' forward from locations with a FormationID, until the next cancelled location, or the next FormationID.
	FormationID         string `xml:"fid,attr"`
	AffectedByDiversion bool   `xml:"affectedByDiversion,attr"`

	// CancellationReason is an optionally provided reason why this location was cancelled.
	CancellationReason *DisruptionReason `xml:"cancelReason"`
}

type OriginLocation struct {
	LocationSchedule
	// PublicArrivalTime is optionally provided.
	PublicArrivalTime railreader.TrainTime `xml:"pta,attr"`
	// PublicDepartureTime is optionally provided.
	PublicDepartureTime railreader.TrainTime `xml:"ptd,attr"`
	// WorkingArrivalTime is optionally provided.
	WorkingArrivalTime   railreader.TrainTime `xml:"wta,attr"`
	WorkingDepartureTime railreader.TrainTime `xml:"wtd,attr"`
	// FalseDestination is an optionally provided destination TIPLOC that is not the train's true destination, but should be displayed to the public as the train's destination, at this location.
	FalseDestination railreader.TIPLOC `xml:"fd,attr"`
}

type OperationalOriginLocation struct {
	LocationSchedule
	// WorkingArrivalTime is optionally provided.
	WorkingArrivalTime   railreader.TrainTime `xml:"wta,attr"`
	WorkingDepartureTime railreader.TrainTime `xml:"wtd,attr"`
}

type IntermediateLocation struct {
	LocationSchedule
	// PublicArrivalTime is optionally provided.
	PublicArrivalTime railreader.TrainTime `xml:"pta,attr"`
	// PublicDepartureTime is optionally provided.
	PublicDepartureTime  railreader.TrainTime `xml:"ptd,attr"`
	WorkingArrivalTime   railreader.TrainTime `xml:"wta,attr"`
	WorkingDepartureTime railreader.TrainTime `xml:"wtd,attr"`
	// RoutingDelay is an optionally provided amount of minutes a change in the train's routing has delayed this location's PublicArrivalTime.
	RoutingDelay int `xml:"rdelay,attr"`
	// FalseDestination is an optionally provided destination TIPLOC that is not the train's true destination, but should be displayed to the public as the train's destination, at this location.
	FalseDestination railreader.TIPLOC `xml:"fd,attr"`
}

type OperationalIntermediateLocation struct {
	LocationSchedule
	WorkingArrivalTime   railreader.TrainTime `xml:"wta,attr"`
	WorkingDepartureTime railreader.TrainTime `xml:"wtd,attr"`
	// RoutingDelay is an optionally provided amount of minutes a change in the train's routing has delayed this location's PublicArrivalTime.
	RoutingDelay int `xml:"rdelay,attr"`
}

type IntermediatePassingLocation struct {
	LocationSchedule
	WorkingPassingTime railreader.TrainTime `xml:"wtp,attr"`
	// RoutingDelay is an optionally provided amount of minutes a change in the train's routing has delayed this location's PublicArrivalTime.
	RoutingDelay int `xml:"rdelay,attr"`
}

type DestinationLocation struct {
	LocationSchedule
	// PublicArrivalTime is optionally provided.
	PublicArrivalTime railreader.TrainTime `xml:"pta,attr"`
	// PublicDepartureTime is optionally provided.
	PublicDepartureTime railreader.TrainTime `xml:"ptd,attr"`
	WorkingArrivalTime  railreader.TrainTime `xml:"wta,attr"`
	// WorkingDepartureTime is optionally provided.
	WorkingDepartureTime railreader.TrainTime `xml:"wtd,attr"`
	// RoutingDelay is an optionally provided amount of minutes a change in the train's routing has delayed this location's PublicArrivalTime.
	RoutingDelay int `xml:"rdelay,attr"`
}

type OperationalDestinationLocation struct {
	LocationSchedule
	WorkingArrivalTime railreader.TrainTime `xml:"wta,attr"`
	// WorkingDepartureTime is optionally provided.
	WorkingDepartureTime railreader.TrainTime `xml:"wtd,attr"`
	// RoutingDelay is an optionally provided amount of minutes a change in the train's routing has delayed this location's PublicArrivalTime.
	RoutingDelay int `xml:"rdelay,attr"`
}
