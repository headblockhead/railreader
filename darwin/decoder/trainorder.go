package decoder

import "github.com/headblockhead/railreader"

// TrainOrder is the order that services are expected to depart from a platform.
// This is most useful for services that divide.
type TrainOrder struct {
	// TIPLOC is the location code for the station with the platform.
	TIPLOC railreader.TIPLOC `xml:"tiploc,attr"`
	// CRS is the passenger-facing 3-letter 'CRS code' for this station.
	CRS      string `xml:"crs,attr"`
	Platform string `xml:"platform,attr"`
	// Clear is true when the current train order should be cleared from the platform.
	Clear    bool               `xml:"clear,attr"`
	Services TrainOrderServices `xml:"set"`
}

type TrainOrderServices struct {
	First  TrainOrderService  `xml:"first"`
	Second TrainOrderService  `xml:"second"`
	Third  *TrainOrderService `xml:"third"`
}

// TrainOrderService can describe a service by RID + time(s), or by its headcode (TrainID).
type TrainOrderService struct {
	RID struct {
		RID string `xml:",chardata"`

		// at least one of:
		WorkingArrivalTime   *railreader.TrainTime `xml:"wta,attr"`
		WorkingDepartureTime *railreader.TrainTime `xml:"wtd,attr"`
		WorkingPassingTime   *railreader.TrainTime `xml:"wtp,attr"`
		PublicArrivalTime    *railreader.TrainTime `xml:"pta,attr"`
		PublicDepartureTime  *railreader.TrainTime `xml:"ptd,attr"`
	} `xml:"rid"`
	TrainID string `xml:"trainID,attr"`
}
