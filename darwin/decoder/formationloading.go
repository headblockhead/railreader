package decoder

import "github.com/headblockhead/railreader"

// FormationLoading contains the estimated percentage loading per coach for a specific formation, location, and service.
type FormationLoading struct {
	// FormationID is used to link this information to a specific train formation.
	FormationID string `xml:"fid,attr"`
	// RID is the unique 16-character ID for a specific train.
	RID string `xml:"rid,attr"`
	// TIPLOC is the code for the location where the loading information applies.
	TIPLOC railreader.TIPLOC `xml:"tpl,attr"`

	// at least one of:
	PublicArrivalTime    railreader.TrainTime `xml:"pta,attr"`
	PublicDepartureTime  railreader.TrainTime `xml:"ptd,attr"`
	WorkingArrivalTime   railreader.TrainTime `xml:"wta,attr"`
	WorkingDepartureTime railreader.TrainTime `xml:"wtd,attr"`
	WorkingPassingTime   railreader.TrainTime `xml:"wtp,attr"`

	Loading []CoachLoadingData `xml:"loading"`
}

type CoachLoadingData struct {
	// CoachIdentifier is the public readable identifier of the coach (eg "A", "B", "1", "2", etc.)
	CoachIdentifier string `xml:"coachNumber,attr"`
	// Source is optional.
	Source string `xml:"src,attr"`
	// SourceSystem is optional. If Source is "CIS", it is most likely a CISCode.
	SourceSystem string `xml:"srcInst,attr"`

	Percentage int `xml:",chardata"`
}
