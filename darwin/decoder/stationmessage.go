package decoder

import (
	"html"
)

type StationMessage struct {
	ID       string                 `xml:"id,attr"`
	Category StationMessageCategory `xml:"category,attr"`
	// Severity is usually 0 or 1.
	Severity int `xml:"severity,attr"`
	// Supressed is true if the message should not be shown to the public.
	Supressed bool `xml:"suppress,attr"`

	// Stations is optional. Sometimes, no stations are listed, even when the message refers exclusively to a specific station!
	// It is probably best to use a different API to get this data.
	Stations []struct {
		// CRS is the passenger-facing 3-letter 'CRS code' for this station.
		CRS string `xml:"crs,attr"`
	} `xml:"Station"`

	// All other tags are interpreted as part of the message body.
	Message XHTMLBody `xml:",any"`
}

type StationMessageCategory string

const (
	StationMessageCategoryStation StationMessageCategory = "Station"
	StationMessageCategoryTrain   StationMessageCategory = "Train"
)

// XHTMLBody is a basic HTML structure, containing exclusively <p> and <a> tags.
type XHTMLBody struct {
	InnerXML string `xml:",innerxml"`
}

func (s *XHTMLBody) UnmarshalText(text []byte) error {
	// Decode the XML into HTML.
	// FIXME
	unescaped := html.UnescapeString(string(text))
	s.InnerXML = unescaped
	return nil
}
