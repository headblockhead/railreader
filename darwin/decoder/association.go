package decoder

import "github.com/headblockhead/railreader"

type AssociationInformation struct {
	TIPLOC   railreader.TIPLOC              `xml:"tiploc,attr"`
	Category railreader.AssociationCategory `xml:"category,attr"`
	// Cancelled indicates the association won't happen.
	Cancelled bool `xml:"isCancelled,attr"`
	// Deleted indicates the association no longer exists.
	Deleted bool `xml:"isDeleted,attr"`

	Main       []AssociatedService `xml:"main"`
	Associated []AssociatedService `xml:"assoc"`
}

type AssociatedService struct {
	// RID is a unique 16-character ID for a specific train.
	RID string `xml:"rid,attr"`

	// At least one of these times must be present.
	WorkingArrivalTime   *railreader.TrainTime `xml:"wta,attr"`
	WorkingDepartureTime *railreader.TrainTime `xml:"wtd,attr"`
	WorkingPassingTime   *railreader.TrainTime `xml:"wtp,attr"`
	PublicArrivalTime    *railreader.TrainTime `xml:"pta,attr"`
	PublicDepartureTime  *railreader.TrainTime `xml:"ptd,attr"`
}
