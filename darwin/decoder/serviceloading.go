package decoder

import "github.com/headblockhead/railreader"

// ServiceLoading contains the estimated percentage loading for an entire service at a specific location.
type ServiceLoading struct {
	// RID is the unique 16-character ID for a specific train.
	RID string `xml:"rid,attr"`
	// TIPLOC is the code for the location where the loading information applies.
	TIPLOC railreader.TIPLOC `xml:"tpl,attr"`

	// at least one of:
	PublicArrivalTime    *railreader.TrainTime `xml:"pta,attr"`
	PublicDepartureTime  *railreader.TrainTime `xml:"ptd,attr"`
	WorkingArrivalTime   *railreader.TrainTime `xml:"wta,attr"`
	WorkingDepartureTime *railreader.TrainTime `xml:"wtd,attr"`
	WorkingPassingTime   *railreader.TrainTime `xml:"wtp,attr"`

	// zero or one of:
	LoadingCategory   *LoadingCategory   `xml:"loadingCategory"`
	LoadingPercentage *LoadingPercentage `xml:"loadingPercentage"`
}

type Loading struct {
	Type         *string  `xml:"type,attr"`
	Source       *string  `xml:"src,attr"`
	SourceSystem *CISCode `xml:"srcInst,attr"`
}

type LoadingCategoryCode string

// TODO: load this from the refdata.

type LoadingCategory struct {
	Loading
	Category *LoadingCategoryCode `xml:",chardata"`
}

type LoadingPercentage struct {
	Loading
	Percentage *int `xml:",chardata"`
}
