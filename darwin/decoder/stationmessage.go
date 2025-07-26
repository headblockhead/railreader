package decoder

type StationMessage struct {
	ID       string `xml:"id,attr"`
	Category string `xml:"cat,attr"`
	// Severity is usually 0 or 1.
	Severity int `xml:"sev,attr"`
	// Supressed is true if the message should not be shown to the public.
	Supressed bool `xml:"suppress,attr"`

	// Stations is a list of the CRS codes of stations where the message should be displayed.
	// It can be empty, however if this is the case, the message shouldn't be shown.
	Stations []StationCRS `xml:"Station"`

	// All other tags are interpreted as part of the message body.
	Message XHTMLBody `xml:",any"`
}

type StationCRS struct {
	// CRS is the passenger-facing 3-letter code for this station.
	CRS string `xml:"crs,attr"`
}

type XHTMLBody struct {
	// Content is a basic HTML string, containing only <p> and <a> tags.
	Content string `xml:",innerxml"`
}
