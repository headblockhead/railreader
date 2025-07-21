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
	// Source is the optionally provided source of the formation data.
	Source *string `xml:"src,attr"`
	// SourceSystem is optional.
	SourceSystem *CISCode `xml:"srcInstance,attr"`

	Coaches []Coach `xml:"coaches>coach"`
}

type Coach struct {
	// CoachIdentifier is the public readable identifier of the coach (eg "A", "B", "1", "2", etc.)
	CoachIdentifier string `xml:"coachNumber,attr"`
	// CoachClass is the optionally provided class of the coach (eg "First", "Standard")
	CoachClass *string `xml:"coachClass,attr"`

	Toilet ToiletInformation `xml:"toilet"`
}

type ToiletInformation struct {
	ToiletStatus ToiletStatus `xml:"status,attr"`
	ToiletType   ToiletType   `xml:",chardata"`
}

func (ti *ToiletInformation) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	// Alias type created to avoid recursion.
	type Alias ToiletInformation
	var toiletinfo Alias

	// Set default values
	toiletinfo.ToiletStatus = ToiletStatusInService
	toiletinfo.ToiletType = ToiletTypeUnknown

	if err := d.DecodeElement(&toiletinfo, &start); err != nil {
		return fmt.Errorf("failed to decode ToiletInformation: %w", err)
	}

	// Convert the alias back to the original type
	*ti = ToiletInformation(toiletinfo)

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
	ToiletTypeAvailable  ToiletType = "Standard"
	ToiletTypeAccessible ToiletType = "Accessible"
)

var ToiletAvailabilityStrings = map[ToiletType]string{
	ToiletTypeUnknown:    "unknown",
	ToiletTypeNone:       "none",
	ToiletTypeAvailable:  "standard",
	ToiletTypeAccessible: "accessible",
}

func (tt ToiletType) String() string {
	if str, ok := ToiletAvailabilityStrings[tt]; ok {
		return str
	}
	return "unknown"
}
