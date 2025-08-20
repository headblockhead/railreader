package unmarshaller

import "github.com/headblockhead/railreader"

// FormationLoading contains the real-time estimated percentage loading per coach for a specific formation+location+service.
type FormationLoading struct {
	LocationTimeIdentifiers
	// RID is the unique 16-character ID for a specific train.
	RID string `xml:"rid,attr"`
	// FormationID is used to link this information to a specific train formation.
	FormationID string `xml:"fid,attr"`
	// TIPLOC is the code for the location where the loading information applies.
	TIPLOC railreader.TimingPointLocationCode `xml:"tpl,attr"`

	Loading []CoachLoading `xml:"loading"`
}

type CoachLoading struct {
	// CoachIdentifier is the public readable identifier of the coach (eg "A", "B", "1", "2", etc.)
	CoachIdentifier string `xml:"coachNumber,attr"`
	// Source is optional.
	Source string `xml:"src,attr"`
	// SourceSystem is optional. If Source is "CIS", it is most likely a CISCode.
	SourceSystem string `xml:"srcInst,attr"`

	Percentage int `xml:",chardata"`
}
