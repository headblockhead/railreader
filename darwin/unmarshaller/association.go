package unmarshaller

import "github.com/headblockhead/railreader"

// Association represents a relationship between two services.
type Association struct {
	TIPLOC   string                         `xml:"tiploc,attr"`
	Category railreader.AssociationCategory `xml:"category,attr"`
	// Cancelled indicates the association won't happen.
	Cancelled *bool `xml:"isCancelled,attr"`
	// Deleted indicates the association no longer exists.
	Deleted *bool `xml:"isDeleted,attr"`

	MainService       AssociatedService `xml:"main"`
	AssociatedService AssociatedService `xml:"assoc"`
}

type AssociatedService struct {
	// RID is the unique 16-character ID for a specific train.
	RID string `xml:"rid,attr"`
	LocationTimeIdentifiers
}
