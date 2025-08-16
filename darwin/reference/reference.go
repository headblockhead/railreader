package reference

import (
	"github.com/headblockhead/railreader"
	"github.com/headblockhead/railreader/darwin/decoder"
)

// Reference version 4
type Reference struct {
	Locations                        []Location              `xml:"LocationRef"`
	TrainOperatingCompanies          []TrainOperatingCompany `xml:"TocRef"`
	LateReasons                      []Reason                `xml:"LateRunningReasons>Reason"`
	CancellationReasons              []Reason                `xml:"CancellationReasons>Reason"`
	ViaTexts                         []ViaConditions         `xml:"Via"`
	CustomerInformationSystemSources []CISSource             `xml:"CISSource"`
	LoadingCategories                []LoadingCategory       `xml:"LoadingCategories>category"`
}

type Location struct {
	Location railreader.TimingPointLocationCode `xml:"tpl,attr"`
	// CRS is optional.
	CRS decoder.CRSCode `xml:"crs,attr"`
	// TOC is optional.
	TOC  decoder.TrainOperatingCompanyCode `xml:"toc,attr"`
	Name string                            `xml:"locname,attr"`
}

type TrainOperatingCompany struct {
	ID   decoder.TrainOperatingCompanyCode `xml:"toc,attr"`
	Name string                            `xml:"tocname,attr"`
	// URL is optional.
	URL string `xml:"url,attr"`
}

type Reason struct {
	// ReasonIDs are not unique between late and cancellation reasons.
	ReasonID    int    `xml:"code,attr"`
	Description string `xml:"reasontext,attr"`
}

// ViaConditions provides a set of source+destination+passing locations that will display a 'via' message.
type ViaConditions struct {
	// DisplayAt is the Computerised Reservation System code for the location where the 'via' message will be displayed.
	DisplayAt string `xml:"at,attr"`

	RequiredDestination      railreader.TimingPointLocationCode `xml:"dest,attr"`
	RequiredCallingLocation1 railreader.TimingPointLocationCode `xml:"loc1,attr"`
	// RequiredCallingLocation2 is optionally provided, but if it is provided it must be after RequiredCallingLocation1 in the schedule for the 'via' message to be displayed.
	RequiredCallingLocation2 railreader.TimingPointLocationCode `xml:"loc2,attr"`

	// Text is the message to be displayed.
	Text string `xml:"viatext,attr"`
}

type CISSource struct {
	CIS  decoder.CISCode `xml:"code,attr"`
	Name string          `xml:"name,attr"`
}

type LoadingCategory struct {
	ID string `xml:"Code,attr"`
	// Name is the name of the loading category, eg "Few seats taken".
	Name string `xml:"Name,attr"`
	// TOC is optional. It is unused as of 2025-08-15.
	TOC decoder.TrainOperatingCompanyCode `xml:"Toc,attr"`

	// TypicalDescription should be shown when ServiceLoading.LoadingCategory.Type == LoadingCategoryTypeTypical
	TypicalDescription string `xml:"TypicalDescription"`
	// ExpectedDescription should be shown when ServiceLoading.LoadingCategory.Type == LoadingCategoryTypeExpected
	ExpectedDescription string `xml:"ExpectedDescription"`
	Definition          string `xml:"Definition"`
	// Colour is a hex RGB or RGBA value, eg "#FF0000" or "#FF000080". It is unused as of 2025-08-15.
	Colour string `xml:"Colour"`
	// Image is a filepath to an image that represents the loading category. It is unused as of 2025-08-15.
	Image string `xml:"Image"`
}
