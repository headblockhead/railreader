package decoder

import "github.com/headblockhead/railreader"

// ForecastTime contains a list of updates to predicted and actual train times at locations along a specific train's schedule.
type ForecastTime struct {
	TrainIdentifiers
	// ReverseFormation indicates whether the remaining service after a train divides will run in reverse formation after the dividing location.
	// I don't know why this is here.
	ReverseFormation bool `xml:"isReverseFormation,attr"`

	// LateReason is optional.
	LateReason *DisruptionReason  `xml:"LateReason"`
	Locations  []ForecastLocation `xml:"Location"`
}

type ForecastLocation struct {
	LocationTimeIdentifiers
	// TIPLOC is the code for the location
	TIPLOC railreader.TIPLOC `xml:"tpl,attr"`

	// zero or one of:
	ArrivalData   *ForecastLocationTimeData `xml:"arr"`
	DepartureData *ForecastLocationTimeData `xml:"dep"`
	PassingData   *ForecastLocationTimeData `xml:"pass"`

	LateReason *DisruptionReason `xml:"LateReason"`
	// Uncertainty data may be provided to indicate there is a risk that this service may be disrupted at this location, along with how and why.
	Uncertainty *Uncertainty `xml:"uncertainty"`
	// AffectedBy is expected to contain a National Rail Enquires incident number, to link multiple services disrupted by the same incident together.
	AffectedBy string `xml:"affectedBy"`
	// Length may or may not match the Formation data. If it is 0, it is unknown.
	Length       int               `xml:"length"`
	PlatformData *ForecastPlatform `xml:"plat"`
	// Suppressed indicates that this service should not be shown to users at this location.
	Suppressed bool `xml:"suppr"`
	// DetachesFromFront is true (at a location where train stock is detached) if train stock will be detached from the front of the train at this location, and false if it will be detached from the rear.
	DetachesFromFront bool `xml:"detachFront"`
}

// ForecastLocationTimeData contains the time data for arrival, departure, or passing a location.
type ForecastLocationTimeData struct {
	// EstimatedTime is optional, generated from the public time table (or the Working Time Table if the location does not have public times).
	EstimatedTime railreader.TrainTime `xml:"et,attr"`
	// WorkingTime is optional, generated from the Working Time Table.
	WorkingTime railreader.TrainTime `xml:"wet,attr"`
	// ActualTime is optional, and may not be reported for all locations.
	ActualTime railreader.TrainTime `xml:"at,attr"`
	// ActualTimeRevoked indicates that a previously given 'actual time' was incorrect, and has been replaced by an estimated time.
	ActualTimeRevoked bool `xml:"atRemoved,attr"`
	// ActualTimeSource is the optionally provided source of the Actual Time data, such as "Manual", "GPS", etc.
	ActualTimeSource string `xml:"atClass,attr"`
	// EstimatedTimeMinimum is optional, and indicates the absolute minimum value the estimated time could be.
	EstimatedTimeMinimum railreader.TrainTime `xml:"etmin,attr"`
	// EstimatedTimeUnknown indicates that the forecast for this location has been manually set to "unknown delay".
	// This is usually shown on signage as "Delayed", without a specific time.
	EstimatedTimeUnknown bool `xml:"etUnknown,attr"`
	// Delayed indicates that the forecast for this location is "unknown delay".
	// This is usually shown on signage as "Delayed", without a specific time.
	Delayed bool `xml:"delayed,attr"`
	// Source is the optionally provided source of the time data, such as "Darwin", "CIS", "TRUST", etc.
	Source string `xml:"src,attr"`
	// SourceSystem is optional. If Source is "CIS", it may be a CISCode. If Source is "TRUST", it may be something like "Auto" or "Manu"
	SourceSystem string `xml:"srcInst,attr"`
}

// Uncertainty contains information about a potential future disruption to a service.
type Uncertainty struct {
	// Status indicates the predicted effect of the uncertainty (eg, delay, cancellation, etc).
	// TODO: find examples of Status values.
	Status string `xml:"status,attr"`

	Reason *DisruptionReason `xml:"reason"`
}

// ForecastPlatform provides the platform a train will be at.
type ForecastPlatform struct {
	// Suppressed indicates that the provided platform data should not be shown to the user.
	Suppressed bool `xml:"platsup,attr"`
	// SuppressedByCIS indicates that the platform data should not be shown to the user, and that this was requested manually.
	SuppressedByCIS bool `xml:"cisPlatsup,attr"`
	// Source is the optionally provided source of the platform data.
	Source PlatformDataSource `xml:"platsrc,attr"`
	// Confirmed indicates the platform is almost certain to be correct.
	Confirmed bool `xml:"conf,attr"`

	Platform string `xml:",chardata"`
}

type PlatformDataSource string

const (
	PlatformDataSourcePlanned   PlatformDataSource = "P"
	PlatformDataSourceAutomatic PlatformDataSource = "A"
	PlatformDataSourceManual    PlatformDataSource = "M"
)

var PlatformDataSourceStrings = map[PlatformDataSource]string{
	PlatformDataSourcePlanned:   "Planned",
	PlatformDataSourceAutomatic: "Automatic",
	PlatformDataSourceManual:    "Manual",
}

func (p PlatformDataSource) String() string {
	if str, ok := PlatformDataSourceStrings[p]; ok {
		return str
	}
	return string(p)
}
