package darwin

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

// PushPortMessage is the root node of Darwin messages.
type PushPortMessage struct {
	Timestamp time.Time `xml:"ts,attr"`
	Version   string    `xml:"version,attr"`

	UpdateResponse   *Response `xml:"uR"`
	SnapshotResponse *Response `xml:"sR"`
}

// Response to a request made to update Darwin's data, by broadcasting the new state(s) of the data to all subscribers.
type Response struct {
	// UpdateOrigin is optionally provided by the requestor to indicate which system the update originated from.
	UpdateOrigin string `xml:"updateOrigin,attr"`
	// RequestSource is optionally provided by the requestor to indicate who they are.
	RequestSource string `xml:"requestSource,attr"`
	// RequestID is optionally provided by the requestor to identify their request.
	RequestID string `xml:"requestID,attr"`

	// 0 or more of any of these updated elements can be present in a response.
	// This includes 0 of all, which is a valid response.

	Schedules                          []ScheduleInformation                `xml:"schedule"`
	Deactivations                      []DeactivationInformation            `xml:"deactivated"`
	Associations                       []AssociationInformation             `xml:"association"`
	TrainFormations                    []TrainFormationInformation          `xml:"scheduleFormations"`
	ActualAndForecasts                 []ActualAndForecastInformation       `xml:"TS"`
	TrainLoadings                      []TrainLoadingInformation            `xml:"formationLoadings"`
	TableSuppressionAndStationMessages []TableSuppressionAndStationMessages `xml:"OW"`
	TrainOrders                        []TrainOrderInformation              `xml:"trainOrder"`
	TrainAlertMessages                 []TrainAlertMessages                 `xml:"trainAlert"`
	TrackingIDChanges                  []TrackingIDChanges                  `xml:"trackingID"`
	Alarms                             []Alarms                             `xml:"alarm"`
}

type ScheduleInformation struct {
	// Locations is a slice of at least 2 location elements that describe the train's schedule.
	Locations []LocationGeneric `xml:",any"`
	// CancellationReasons are the reasons why this service was cancelled.
	// This is provided at the service level, and/or the location level.
	CancellationReasons []DisruptionReason `xml:"cancelReason"`
	// DivertedVias are the TIPLOCs that this service is diverted via.
	DivertedVias []TIPLOC `xml:"divertedVia"`
	// DiversionReasons are the reasons why this service is diverted.
	DiversionReasons []DisruptionReason `xml:"diversionReason"`

	// RID is the unique 16-character ID for a specific train, running this schedule, at this time.
	RID string `xml:"rid,attr"`
	// UID is (despite the name) a non-unique 6-character ID for this route at this time of day.
	UID string `xml:"uid,attr"`
	// TrainID is the 4-character headcode of the train, with the format:
	// [0-9][A-Z][0-9][0-9]
	TrainID string `xml:"trainID,attr"`
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
	Status ServiceType `xml:"status,attr"`
	// TrainCategory is a 2-character code for the type of train.
	// Values that indicate a passenger service are:
	// OL, OO, OW, XC, XD, XI, XR, XX, XZ.
	// TODO: provide enum or string func for these. (see CIF)
	// If not provided, it defaults to OO.
	TrainCategory string `xml:"trainCat,attr"`
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
	schedule.Status = PassengerAndParcelTrain
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

// Location is the base struct for all location types.
type Location struct {
	// TIPLOC is the code for the location
	TIPLOC TIPLOC `xml:"tpl,attr"`
	// Activity optionally provides what is happening at this location.
	Activity ActivityCode `xml:"act,attr"`
	// PlannedActivity optionally provides what was/is planned to happen at this location.
	// This is only usually given if the Activity is different to the PlannedActivity.
	PlannedActivity ActivityCode `xml:"planAct,attr"`
	Cancelled       bool         `xml:"can,attr"`
	// FormationID is the ID of the train formation that is used at this location.
	// Formations 'ripple' forward from locations with a FormationID, until the next cancelled location, or the next FormationID.
	FormationID         string `xml:"fid,attr"`
	AffectedByDiversion bool   `xml:"affectedByVersion,attr"`
	// CancellationReasons is an optionally provided list of reasons why this location was cancelled. A reason may be provided per location, but may also be provided for the entire service.
	CancellationReasons []DisruptionReason `xml:"cancelreason"`
}

type OriginLocation struct {
	Location
	// PublicArrivalTime is optionally provided.
	PublicArrivalTime TrainTime `xml:"pta,attr"`
	// PublicDepartureTime is optionally provided.
	PublicDepartureTime TrainTime `xml:"ptd,attr"`
	// WorkingArrivalTime is optionally provided.
	WorkingArrivalTime   TrainTime `xml:"wta,attr"`
	WorkingDepartureTime TrainTime `xml:"wtd,attr"`
	// FalseDestination is an optionally provided destination TIPLOC that is not the train's true destination, but should be displayed to the public as the train's destination, at this location.
	FalseDestination TIPLOC `xml:"fd,attr"`
}

type OperationalOriginLocation struct {
	Location
	// WorkingArrivalTime is optionally provided.
	WorkingArrivalTime   TrainTime `xml:"wta,attr"`
	WorkingDepartureTime TrainTime `xml:"wtd,attr"`
}

type IntermediateLocation struct {
	Location
	// PublicArrivalTime is optionally provided.
	PublicArrivalTime TrainTime `xml:"pta,attr"`
	// PublicDepartureTime is optionally provided.
	PublicDepartureTime  TrainTime `xml:"ptd,attr"`
	WorkingArrivalTime   TrainTime `xml:"wta,attr"`
	WorkingDepartureTime TrainTime `xml:"wtd,attr"`
	// TODO: confirm this!
	// RoutingDelay is an optionally provided amount of minutes that the train's arrival has been adjusted by at this location (positive or negative) due to routing changes.
	RoutingDelay int `xml:"rdelay,attr"`
}

type OperationalIntermediateLocation struct {
	Location
	WorkingArrivalTime   TrainTime `xml:"wta,attr"`
	WorkingDepartureTime TrainTime `xml:"wtd,attr"`
	// TODO: confirm this!
	// RoutingDelay is an optionally provided amount of minutes that the train's arrival has been adjusted by at this location (positive or negative) due to routing changes.
	RoutingDelay int `xml:"rdelay,attr"`
}

type IntermediatePassingLocation struct {
	Location
	WorkingPassingTime TrainTime `xml:"wtp,attr"`
	// TODO: confirm this!
	// RoutingDelay is an optionally provided amount of minutes that the train's arrival has been adjusted by at this location (positive or negative) due to routing changes.
	RoutingDelay int `xml:"rdelay,attr"`
}

type DestinationLocation struct {
	Location
	// PublicArrivalTime is optionally provided.
	PublicArrivalTime TrainTime `xml:"pta,attr"`
	// PublicDepartureTime is optionally provided.
	PublicDepartureTime TrainTime `xml:"ptd,attr"`
	WorkingArrivalTime  TrainTime `xml:"wta,attr"`
	// WorkingDepartureTime is optionally provided.
	WorkingDepartureTime TrainTime `xml:"wtd,attr"`
	// TODO: confirm this!
	// RoutingDelay is an optionally provided amount of minutes that the train's arrival has been adjusted by at this location (positive or negative) due to routing changes.
	RoutingDelay int `xml:"rdelay,attr"`
}

type OperationalDestinationLocation struct {
	Location
	WorkingArrivalTime TrainTime `xml:"wta,attr"`
	// WorkingDepartureTime is optionally provided.
	WorkingDepartureTime TrainTime `xml:"wtd,attr"`
	// TODO: confirm this!
	// RoutingDelay is an optionally provided amount of minutes that the train's arrival has been adjusted by at this location (positive or negative) due to routing changes.
	RoutingDelay int `xml:"rdelay,attr"`
}

// DeactivationInformation is sent to indicate a RID is expected to recieve no further updates, and shouldn't be displayed publicly.
// A deactivation can be un-done by a subsequent ScheduleInformation with the same RID.
type DeactivationInformation struct {
	// RID is the unique 16-character ID for the specific train+schedule+time combo that has been deactivated.
	RID string `xml:"rid,attr"`
}

type AssociationInformation struct {
}

type TrainFormationInformation struct {
}

type ActualAndForecastInformation struct {
}

type TrainLoadingInformation struct {
}

type TableSuppressionAndStationMessages struct {
}

type TrainOrderInformation struct {
}

type TrainAlertMessages struct {
}

type TrackingIDChanges struct {
}

type Alarms struct {
}

func (dc *Connection) ProcessMessageCapsule(msg MessageCapsule) error {
	log := dc.log.With(slog.String("messageID", string(msg.MessageID)))

	os.WriteFile(filepath.Join("capture", msg.MessageID+".xml"), []byte(msg.Bytes), 0644)

	var pport PushPortMessage
	if err := xml.Unmarshal([]byte(msg.Bytes), &pport); err != nil {
		return fmt.Errorf("failed to unmarshal message XML: %w", err)
	}

	// TODO: check common fields are always as we expect

	js, err := json.Marshal(pport)
	if err != nil {
		return err
	}

	if pport.UpdateResponse != nil && pport.UpdateResponse.Schedules != nil {
		log.Info("unmarshaled", slog.String("msg", string(js)))
	}

	return nil
}
