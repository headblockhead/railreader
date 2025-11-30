package unmarshaller

import (
	"encoding/xml"
	"fmt"
)

// ServiceLoading contains the typical percentage loading (or LoadingCategory) for an entire service at a specific location. It does not vary based on real-time data.
type ServiceLoading struct {
	LocationTimeIdentifiers
	// RID is the unique 16-character ID for a specific train.
	RID string `xml:"rid,attr"`
	// TIPLOC is the code for the location where the loading information applies.
	TIPLOC string `xml:"tpl,attr"`

	// zero or one of:
	LoadingCategory   *LoadingCategory   `xml:"loadingCategory"`
	LoadingPercentage *LoadingPercentage `xml:"loadingPercentage"`
}

type LoadingCategory struct {
	// Type can be "Expected" or "Typical", and defaults to "Typical" if not specified.
	Type LoadingCategoryType `xml:"type,attr"`
	// Source is optional.
	Source *string `xml:"src,attr"`
	// SourceSystem is optional. If Source is "CIS", it is most likely a CISCode.
	SourceSystem *string `xml:"srcInst,attr"`

	// Category is between 1 and 4 characters, and can be looked up in the reference data.
	Category string `xml:",chardata"`
}

type LoadingCategoryType string

const (
	LoadingCategoryTypeExpected LoadingCategoryType = "Expected"
	LoadingCategoryTypeTypical  LoadingCategoryType = "Typical"
)

func (lc *LoadingCategory) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	// Alias type created to avoid recursion.
	type Alias LoadingCategory
	var loadingCategory Alias

	// Set default values.
	loadingCategory.Type = LoadingCategoryTypeTypical

	if err := d.DecodeElement(&loadingCategory, &start); err != nil {
		return fmt.Errorf("failed to decode LoadingCategory: %w", err)
	}

	// Convert the alias back to the original type.
	*lc = LoadingCategory(loadingCategory)

	return nil
}

type LoadingPercentage struct {
	// Type defaults to "Typical" if not specified.
	Type string `xml:"type,attr"`
	// Source is optional.
	Source *string `xml:"src,attr"`
	// SourceSystem is optional. If Source is "CIS", it is most likely a CISCode.
	SourceSystem *string `xml:"srcInst,attr"`

	// Percentage is between 0 and 100, inclusive.
	Percentage int `xml:",chardata"`
}

func (lp *LoadingPercentage) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	// Alias type created to avoid recursion.
	type Alias LoadingPercentage
	var loadingPercentage Alias

	// Set default values.
	loadingPercentage.Type = "Typical"

	if err := d.DecodeElement(&loadingPercentage, &start); err != nil {
		return fmt.Errorf("failed to decode LoadingPercentage: %w", err)
	}

	// Convert the alias back to the original type.
	*lp = LoadingPercentage(loadingPercentage)

	return nil
}
