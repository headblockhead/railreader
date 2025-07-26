package decoder

import "github.com/headblockhead/railreader"

// TrainOrder is the order that services are expected to depart from a platform.
// This is most useful for services that divide.
type TrainOrder struct {
	// TIPLOC is the location code for the station with the platform.
	TIPLOC railreader.TIPLOC `xml:"tiploc,attr"`
	// CRS is the passenger-facing 3-letter code for this station.
	CRS      string `xml:"crs,attr"`
	Platform string `xml:"platform,attr"`

	// only one of:
	// ClearOrder is true when the current train order should be cleared from the platform.
	ClearOrder trueIfPresent      `xml:"clear"`
	Services   TrainOrderServices `xml:"set"`
}

type TrainOrderServices struct {
	First  TrainOrderService `xml:"first"`
	Second TrainOrderService `xml:"second"`
	// Third is optional.
	Third *TrainOrderService `xml:"third"`
}

// TrainOrderService can describe a service at a specific point on its route by RID, time(s), and a seperatly provided TIPLOC,
// or (if it is not included in Darwin) by its headcode.
type TrainOrderService struct {
	// only one of:
	RIDAndTime OrderedService `xml:"rid"`
	Headcode   string         `xml:"trainID"`
}

type OrderedService struct {
	LocationTimeIdentifiers
	// RID is a unique 16-character ID for a specific train.
	RID string `xml:",chardata"`
}
