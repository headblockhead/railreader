package unmarshaller

import (
	"encoding/xml"
)

// Reference version 4
type Reference struct {
	ID                         string                           `xml:"timetableID,attr"`
	Locations                  []LocationReference              `xml:"LocationRef"`
	TrainOperatingCompanies    []TrainOperatingCompanyReference `xml:"TocRef"`
	LateReasons                []ReasonDescription              `xml:"LateRunningReasons>Reason"`
	CancellationReasons        []ReasonDescription              `xml:"CancellationReasons>Reason"`
	ViaConditions              []ViaCondition                   `xml:"Via"`
	CustomerInformationSystems []CISReference                   `xml:"CISSource"`
	LoadingCategories          []LoadingCategoryReference       `xml:"LoadingCategories>category"`
}

func NewReference(xmlData string) (ref Reference, err error) {
	if err = xml.Unmarshal([]byte(xmlData), &ref); err != nil {
		return
	}
	return
}

type LocationReference struct {
	Location string `xml:"tpl,attr"`
	// CRS is optional.
	CRS *string `xml:"crs,attr"`
	// TOC is optional.
	TOC  *string `xml:"toc,attr"`
	Name string  `xml:"locname,attr"`
}

type TrainOperatingCompanyReference struct {
	ID   string `xml:"toc,attr"`
	Name string `xml:"tocname,attr"`
	// URL is optional.
	URL *string `xml:"url,attr"`
}

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

	RequiredDestination      string `xml:"dest,attr"`
	RequiredCallingLocation1 string `xml:"loc1,attr"`
	// RequiredCallingLocation2 is optionally provided.
	// If it is provided, it must be after RequiredCallingLocation1 in the schedule for the 'via' message to be displayed.
	RequiredCallingLocation2 *string `xml:"loc2,attr"`

	// Text to be displayed when the conditions are met.
	Text string `xml:"viatext,attr"`
}

type CISReference struct {
	CIS  string `xml:"code,attr"`
	Name string `xml:"name,attr"`
}

type LoadingCategoryReference struct {
	Code string `xml:"Code,attr"`
	// When TOC is not provided, the loading category data for this Code is the default for all TOCs.
	// When TOC is provided, the loading category data here only applies to the specified TOC, and this TOC-specific data should override the default.
	TOC *string `xml:"Toc,attr"`
	// Name is a short display name of the loading category, eg "Few seats taken".
	Name string `xml:"Name,attr"`

	// TypicalDescription should be shown when ServiceLoading.LoadingCategory.Type == LoadingCategoryTypeTypical
	TypicalDescription string `xml:"TypicalDescription"`
	// ExpectedDescription should be shown when ServiceLoading.LoadingCategory.Type == LoadingCategoryTypeExpected
	ExpectedDescription string `xml:"ExpectedDescription"`
	// Definition is a longer description of the loading category that defines more specifically what it means.
	Definition string `xml:"Definition"`
	// Colour is a hex RGB or RGBA value, eg "#FF0000" or "#FF000080".
	Colour string `xml:"Colour"`
	// Image is a filepath to an icon that represents the loading category.
	Image string `xml:"Image"`
}
