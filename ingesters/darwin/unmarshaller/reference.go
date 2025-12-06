package unmarshaller

import (
	"encoding/xml"
	"io"
)

const ExpectedReferenceFileSuffix = "_ref_v4.xml.gz"

// Reference version 4
type Reference struct {
	ID string `xml:"timetableId,attr"`
	// All possible TIming Point LOCcations.
	Locations []LocationReference `xml:"LocationRef"`
	// Most (but not all) Train Operating Companies.
	TrainOperatingCompanies []TrainOperatingCompanyReference `xml:"TocRef"`
	// All possible reasons for late running.
	LateReasons []ReasonDescription `xml:"LateRunningReasons>Reason"`
	// All possible Reasons for cancellation.
	CancellationReasons []ReasonDescription `xml:"CancellationReasons>Reason"`
	// Conditions that must be met for a 'via' message to be displayed for a service.
	ViaConditions []ViaCondition `xml:"Via"`
	// Most (but not all) of the CustomerInformationSystems and their names.
	CustomerInformationSystems []CISReference `xml:"CISSource"`
	// Categories that can be used to represent a rough estimate on how full a train is.
	LoadingCategories []LoadingCategoryReference `xml:"LoadingCategories>category"`
}

func NewReference(xmlData io.Reader) (*Reference, error) {
	decoder := xml.NewDecoder(xmlData)
	var ref Reference
	if err := decoder.Decode(&ref); err != nil {
		return nil, err
	}
	return &ref, nil
}

// LocationReference holds data about a timing point location.
type LocationReference struct {
	// TIPLOC code (variable-length code representing a specific timing point).
	Location string `xml:"tpl,attr"`
	// (optional) Computerised Reservation System ID (3-letter code identifying a specific passenger station).
	CRS *string `xml:"crs,attr"`
	// (optional) Train Operating Company that operates the location
	TOC  *string `xml:"toc,attr"`
	Name string  `xml:"locname,attr"`
}

// TrainOperatingCompanyReference holds data about a Train Operating Company.
type TrainOperatingCompanyReference struct {
	ID   string `xml:"toc,attr"`
	Name string `xml:"tocname,attr"`
	// (optional) URL of (typically a page that will redirect to) the TOC's website.
	URL *string `xml:"url,attr"`
}

// ReasonDescription holds a reason code and description for either late running or cancellations.
// IDs are not unique between late and cancellation reasons, so the context must be known when using them.
type ReasonDescription struct {
	// ReasonIDs are not unique between late and cancellation reasons.
	ReasonID    int    `xml:"code,attr"`
	Description string `xml:"reasontext,attr"`
}

// ViaCondition provides a set of source+destination+passing locations that will display a 'via' message.
// When searching through ViaConditions, the first match should be used,
// and the search must be performed in the same order as they appear in the XML; they are listed in priority order.
type ViaCondition struct {
	// DisplayAt is the Computerised Reservation System code for the location where the 'via' message will be displayed.
	DisplayAt string `xml:"at,attr"`

	// The service must have this location as its destination for the via condition to be met.
	RequiredDestination string `xml:"dest,attr"`
	// The service must call at this location for the via condition to be met.
	RequiredCallingLocation1 string `xml:"loc1,attr"`
	// If specified, the service must call at this location at some point after the first_required_location_id for the via condition to be met.
	RequiredCallingLocation2 *string `xml:"loc2,attr"`

	// Text to be displayed when the conditions are met.
	Text string `xml:"viatext,attr"`
}

// CISReference holds data about a Customer Information System.
type CISReference struct {
	// CIS is the code identifying the CIS.
	CIS string `xml:"code,attr"`
	// Name is the human-readable name of the CIS.
	Name string `xml:"name,attr"`
}

type LoadingCategoryReference struct {
	Code string `xml:"Code,attr"`
	// When TOC is not provided, the loading category data for this Code is the default for all TOCs.
	// When TOC is provided, the loading category data here only applies to the specified TOC, and this TOC-specific data should override the default.
	TOC *string `xml:"Toc,attr"`
	// Name is a short display name of the loading category.
	// Example: "Few seats taken"
	Name string `xml:"Name,attr"`

	// TypicalDescription should be shown when ServiceLoading.LoadingCategory.Type == LoadingCategoryTypeTypical,
	// Example: "Usually, only a few seats taken"
	TypicalDescription string `xml:"TypicalDescription"`
	// ExpectedDescription should be shown when ServiceLoading.LoadingCategory.Type == LoadingCategoryTypeExpected,
	// Example: "Only a few seats taken"
	ExpectedDescription string `xml:"ExpectedDescription"`
	// Definition of what customers can expect.
	// Example: "Everyone will be able to find a seat"
	Definition string `xml:"Definition"`
	// Colour is a hex RGB or RGBA value, eg "#FF0000" or "#FF000080".
	Colour string `xml:"Colour"`
	// Image is a filepath to an icon that represents the loading category.
	Image string `xml:"Image"`
}
