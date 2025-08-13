package decoder

import (
	"encoding/xml"
	"fmt"
)

type FormationsOfService struct {
	// RID is a unique 16-character ID for a specific train.
	RID string `xml:"rid,attr"`

	// Formations is a list of the various formations the train may use during its journey.
	Formations []Formation `xml:"formation"`
}

type Formation struct {
	ID string `xml:"fid,attr"`
	// Source is optional.
	Source string `xml:"src,attr"`
	// SourceSystem is optional. If Source is "CIS", it is most likely a CISCode.
	SourceSystem string `xml:"srcInst,attr"`

	Coaches []FormationCoach `xml:"coaches>coach"`
}

type FormationCoach struct {
	// Identifier is the public readable identifier of the coach (eg "A", "B", "1", "2", etc.)
	Identifier string `xml:"coachNumber,attr"`
	// Class is the optionally provided class of the coach (eg "First", "Standard")
	Class string `xml:"coachClass,attr"`

	Toilet FormationCoachToilet `xml:"toilet"`
}

type FormationCoachToilet struct {
	Status ToiletStatus `xml:"status,attr"`
	Type   ToiletType   `xml:",chardata"`
}

func (ti *FormationCoachToilet) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	// Alias type created to avoid recursion.
	type Alias FormationCoachToilet
	var toiletinfo Alias

	// Set default values.
	toiletinfo.Status = ToiletStatusInService

	if err := d.DecodeElement(&toiletinfo, &start); err != nil {
		return fmt.Errorf("failed to decode ToiletInformation: %w", err)
	}

	// If the ToiletType is empty, set it to the default value.
	if toiletinfo.Type == "" {
		toiletinfo.Type = ToiletTypeUnknown
	}

	// Convert the alias back to the original type.
	*ti = FormationCoachToilet(toiletinfo)

	return nil
}

type ToiletStatus string

const (
	ToiletStatusUnknown      ToiletStatus = "Unknown"
	ToiletStatusInService    ToiletStatus = "InService"
	ToiletStatusNotInService ToiletStatus = "NotInService"
)

var ToiletStatusStrings = map[ToiletStatus]string{
	ToiletStatusUnknown:      "unknown",
	ToiletStatusInService:    "in service",
	ToiletStatusNotInService: "out of service",
}

func (ts ToiletStatus) String() string {
	if str, ok := ToiletStatusStrings[ts]; ok {
		return str
	}
	return "unknown"
}

type ToiletType string

const (
	ToiletTypeUnknown    ToiletType = "Unknown"
	ToiletTypeNone       ToiletType = "None"
	ToiletTypeStandard   ToiletType = "Standard"
	ToiletTypeAccessible ToiletType = "Accessible"
)

var ToiletAvailabilityStrings = map[ToiletType]string{
	ToiletTypeUnknown:    "unknown",
	ToiletTypeNone:       "none",
	ToiletTypeStandard:   "standard",
	ToiletTypeAccessible: "accessible",
}

func (tt ToiletType) String() string {
	if str, ok := ToiletAvailabilityStrings[tt]; ok {
		return str
	}
	return "unknown"
}
