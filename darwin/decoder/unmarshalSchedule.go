package decoder

import (
	"encoding/xml"
	"fmt"

	"github.com/headblockhead/railreader"
)

type ScheduleInformation struct {
	// Locations is a slice of at least 2 location elements that describe the train's schedule.
	Locations []LocationGeneric `xml:",any"` // Any other elements will be interpreted as locations.
	// CancellationReasons are the reasons why this service was cancelled.
	// This is provided at the service level, and/or the location level.
	CancellationReasons []railreader.DisruptionReason `xml:"cancelReason"`
	// DivertedVias are the TIPLOCs that this service is diverted via.
	DivertedVias []railreader.TIPLOC `xml:"divertedVia"`
	// DiversionReasons are the reasons why this service is diverted.
	DiversionReasons []railreader.DisruptionReason `xml:"diversionReason"`

	// RID is the unique 16-character ID for a specific train, running this schedule, at this time.
	RID string `xml:"rid,attr"`
	// UID is (despite the name) a non-unique 6-character ID for this route at this time of day.
	UID string `xml:"uid,attr"`
	// TrainID is the 4-character headcode of the train, with the format:
	// [0-9][A-Z][0-9][0-9]
	TrainID string `xml:"trainId,attr"`
	// RSID is the optionally provided Retail Service ID, either as an:
	// 8 character "portion identifier" (including a leading TOC code),
	// or a 6 character "base identifier" (without a TOC code).
	RSID string `xml:"rsid,attr"`
	// SSD is the scheduled start date of the train, in YYYY-MM-DD format.
	SSD string `xml:"ssd,attr"`
	// TOC is the Rail Delivery Group's 2-character code for the train operating company.
	TOC string `xml:"toc,attr"`
	// Status is the 1-character code for the type of transport.
	// If not provided, it defaults to P (Passenger and Parcel Train).
	Status railreader.ServiceType `xml:"status,attr"`
	// TrainCategory is a 2-character code for the type of train.
	// Values that indicate a passenger service are:
	// OL, OO, OW, XC, XD, XI, XR, XX, XZ.
	// TODO: provide enum or string func for these. (see CIF)
	// If not provided, it defaults to OO.
	TrainCategory railreader.TrainCategory `xml:"trainCat,attr"`
	// IsPassengerService is true if not provided. This will sometimes be false, based on the value of the TrainCategory.
	IsPassengerService bool `xml:"isPassengerSvc,attr"`
	// IsActive is true if not provided. It is only present in snapshots, used to indicate a service has been deactivated by a DeactivationInformation element.
	IsActive bool `xml:"isActive,attr"`
	// Deleted means you should not use or display this schedule.
	Deleted   bool `xml:"deleted,attr"`
	IsCharter bool `xml:"isCharter,attr"`
}

func (si *ScheduleInformation) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	// Alias type created used to avoid recursion.
	type Alias ScheduleInformation
	var schedule Alias

	// Set default values
	schedule.Status = railreader.PassengerAndParcelTrain
	schedule.TrainCategory = "OO"
	schedule.IsPassengerService = true
	schedule.IsActive = true

	if err := d.DecodeElement(&schedule, &start); err != nil {
		return fmt.Errorf("failed to decode ScheduleInformation: %w", err)
	}

	// Convert the alias back to the original type
	*si = ScheduleInformation(schedule)

	return nil
}

// LocationGeneric is a generic struct that contains (nullable pointers to) all the possible location types.
type LocationGeneric struct {
	LocationType                    string
	OriginLocation                  *OriginLocation                  `xml:"OR"`
	OperationalOriginLocation       *OperationalOriginLocation       `xml:"OPOR"`
	IntermediateLocation            *IntermediateLocation            `xml:"IP"`
	OperationalIntermediateLocation *OperationalIntermediateLocation `xml:"OPIP"`
	IntermediatePassingLocation     *IntermediatePassingLocation     `xml:"PP"`
	DestinationLocation             *DestinationLocation             `xml:"DT"`
	OperationalDestinationLocation  *OperationalDestinationLocation  `xml:"OPDT"`
}

func (lg *LocationGeneric) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	locationType := start.Name.Local
	lg.LocationType = locationType
	switch locationType {
	case "OR":
		lg.OriginLocation = &OriginLocation{Location: Location{}}
		if err := d.DecodeElement(lg.OriginLocation, &start); err != nil {
			return fmt.Errorf("failed to decode OriginLocation: %w", err)
		}
	case "OPOR":
		lg.OperationalOriginLocation = &OperationalOriginLocation{Location: Location{}}
		if err := d.DecodeElement(lg.OperationalOriginLocation, &start); err != nil {
			return fmt.Errorf("failed to decode OperationalOriginLocation: %w", err)
		}
	case "IP":
		lg.IntermediateLocation = &IntermediateLocation{Location: Location{}}
		if err := d.DecodeElement(lg.IntermediateLocation, &start); err != nil {
			return fmt.Errorf("failed to decode IntermediateLocation: %w", err)
		}
	case "OPIP":
		lg.OperationalIntermediateLocation = &OperationalIntermediateLocation{Location: Location{}}
		if err := d.DecodeElement(lg.OperationalIntermediateLocation, &start); err != nil {
			return fmt.Errorf("failed to decode OperationalIntermediateLocation: %w", err)
		}
	case "PP":
		lg.IntermediatePassingLocation = &IntermediatePassingLocation{Location: Location{}}
		if err := d.DecodeElement(lg.IntermediatePassingLocation, &start); err != nil {
			return fmt.Errorf("failed to decode IntermediatePassingLocation: %w", err)
		}
	case "DT":
		lg.DestinationLocation = &DestinationLocation{Location: Location{}}
		if err := d.DecodeElement(lg.DestinationLocation, &start); err != nil {
			return fmt.Errorf("failed to decode DestinationLocation: %w", err)
		}
	case "OPDT":
		lg.OperationalDestinationLocation = &OperationalDestinationLocation{Location: Location{}}
		if err := d.DecodeElement(lg.OperationalDestinationLocation, &start); err != nil {
			return fmt.Errorf("failed to decode OperationalDestinationLocation: %w", err)
		}
	default:
		return fmt.Errorf("unknown location type: %s", locationType)
	}

	return nil
}

func (lg LocationGeneric) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	switch lg.LocationType {
	case "OR":
		if lg.OriginLocation != nil {
			return e.EncodeElement(lg.OriginLocation, start)
		}
	case "OPOR":
		if lg.OperationalOriginLocation != nil {
			return e.EncodeElement(lg.OperationalOriginLocation, start)
		}
	case "IP":
		if lg.IntermediateLocation != nil {
			return e.EncodeElement(lg.IntermediateLocation, start)
		}
	case "OPIP":
		if lg.OperationalIntermediateLocation != nil {
			return e.EncodeElement(lg.OperationalIntermediateLocation, start)
		}
	case "PP":
		if lg.IntermediatePassingLocation != nil {
			return e.EncodeElement(lg.IntermediatePassingLocation, start)
		}
	case "DT":
		if lg.DestinationLocation != nil {
			return e.EncodeElement(lg.DestinationLocation, start)
		}
	case "OPDT":
		if lg.OperationalDestinationLocation != nil {
			return e.EncodeElement(lg.OperationalDestinationLocation, start)
		}
	default:
		return fmt.Errorf("unknown location type: %s", lg.LocationType)
	}

	return nil
}

// Location is the base struct for all location types.
type Location struct {
	// TIPLOC is the code for the location
	TIPLOC railreader.TIPLOC `xml:"tpl,attr"`
	// Activity optionally provides what is happening at this location.
	Activity railreader.ActivityCode `xml:"act,attr"`
	// PlannedActivity optionally provides what was/is planned to happen at this location.
	// This is only usually given if the Activity is different to the PlannedActivity.
	PlannedActivity railreader.ActivityCode `xml:"planAct,attr"`
	Cancelled       bool                    `xml:"can,attr"`
	// FormationID is the ID of the train formation that is used at this location.
	// Formations 'ripple' forward from locations with a FormationID, until the next cancelled location, or the next FormationID.
	FormationID         string `xml:"fid,attr"`
	AffectedByDiversion bool   `xml:"affectedByVersion,attr"`
	// CancellationReasons is an optionally provided list of reasons why this location was cancelled. A reason may be provided per location, but may also be provided for the entire service.
	CancellationReasons []railreader.DisruptionReason `xml:"cancelreason"`
}

type OriginLocation struct {
	Location
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
	Location
	// WorkingArrivalTime is optionally provided.
	WorkingArrivalTime   railreader.TrainTime `xml:"wta,attr"`
	WorkingDepartureTime railreader.TrainTime `xml:"wtd,attr"`
}

type IntermediateLocation struct {
	Location
	// PublicArrivalTime is optionally provided.
	PublicArrivalTime railreader.TrainTime `xml:"pta,attr"`
	// PublicDepartureTime is optionally provided.
	PublicDepartureTime  railreader.TrainTime `xml:"ptd,attr"`
	WorkingArrivalTime   railreader.TrainTime `xml:"wta,attr"`
	WorkingDepartureTime railreader.TrainTime `xml:"wtd,attr"`
	// RoutingDelay is an optionally provided amount of minutes a change in the train's routing has delayed this location's PublicArrivalTime.
	RoutingDelay int `xml:"rdelay,attr"`
}

type OperationalIntermediateLocation struct {
	Location
	WorkingArrivalTime   railreader.TrainTime `xml:"wta,attr"`
	WorkingDepartureTime railreader.TrainTime `xml:"wtd,attr"`
	// RoutingDelay is an optionally provided amount of minutes a change in the train's routing has delayed this location's PublicArrivalTime.
	RoutingDelay int `xml:"rdelay,attr"`
}

type IntermediatePassingLocation struct {
	Location
	WorkingPassingTime railreader.TrainTime `xml:"wtp,attr"`
	// RoutingDelay is an optionally provided amount of minutes a change in the train's routing has delayed this location's PublicArrivalTime.
	RoutingDelay int `xml:"rdelay,attr"`
}

type DestinationLocation struct {
	Location
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
	Location
	WorkingArrivalTime railreader.TrainTime `xml:"wta,attr"`
	// WorkingDepartureTime is optionally provided.
	WorkingDepartureTime railreader.TrainTime `xml:"wtd,attr"`
	// RoutingDelay is an optionally provided amount of minutes a change in the train's routing has delayed this location's PublicArrivalTime.
	RoutingDelay int `xml:"rdelay,attr"`
}
