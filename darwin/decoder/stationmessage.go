package decoder

type StationMessage struct {
	ID       string `xml:"id,attr"`
	Category string `xml:"cat,attr"`
	// Severity is usually 0 or 1.
	Severity int `xml:"sev,attr"`
	// Supressed is true if the message should not be shown to the public.
	Supressed bool `xml:"suppress,attr"`

	// Stations is a list of the CRS codes of stations where the message should be displayed.
	// It is optional, however if no stations are listed the message won't (& shouldn't) be shown.
	Stations []StationCRS `xml:"Station"`

	// All other tags are interpreted as part of the message body.
	Message XHTMLBody `xml:",any"`
}

type XHTMLBody struct {
	Content string `xml:",innerxml"`
}

type StationCRS struct {
	// CRS is the passenger-facing 3-letter code for this station.
	CRS string `xml:"crs,attr"`
}
