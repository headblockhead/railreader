package unmarshaller

type StationMessage struct {
	ID       string                 `xml:"id,attr"`
	Category StationMessageCategory `xml:"cat,attr"`
	// Severity is usually 0 or 1.
	Severity StationMessageSeverity `xml:"sev,attr"`
	// Supressed is true if the message should not be shown to the public.
	Supressed bool `xml:"suppress,attr"`

	// Stations is a list of the CRS codes of stations where the message should be displayed.
	// It can be empty, however if this is the case, the message shouldn't be shown.
	Stations []StationCRS `xml:"Station"`

	// All other tags are interpreted as part of the message body.
	Message XHTMLBody `xml:",any"`
}

type StationCRS struct {
	CRS CRSCode `xml:"crs,attr"`
}

type XHTMLBody struct {
	// Content is a basic HTML string, containing only <p> and <a> tags.
	// WARNING: This is not output in the same format as TrainAlerts, character entities inside of elements (like &nbsp;) ARE NOT decoded! See stationmessage_test.go for an example.
	Content string `xml:",innerxml"`
}

type StationMessageCategory string

const (
	// StationMessageCategoryStation is about the station itself, such as about lifts, escalators, etc.
	StationMessageCategoryStation StationMessageCategory = "Station"
	// StationMessageCategoryTrain is about the trains that call at a station.
	StationMessageCategoryTrain StationMessageCategory = "Train"
	// StationMessageCategoryPriorTrain is an advance message about something that will affect trains in the future.
	StationMessageCategoryPriorTrain StationMessageCategory = "PriorTrains"
	// StationMessageCategoryPriorOther is an advance message about something in the future, such as lifts being out of order for the next week.
	StationMessageCategoryPriorOther StationMessageCategory = "PriorOthers"
	// StationMessageCategoryConnections is about the connecting services at a station, such as the London Underground.
	StationMessageCategoryConnections StationMessageCategory = "Connections"
	// StationMessageCategorySystem is about the Darwin system itself.
	StationMessageCategorySystem StationMessageCategory = "System"
	// StationMessageCategoryMiscellaneous is for any other messages that don't fit into the above categories.
	StationMessageCategoryMiscellaneous StationMessageCategory = "Misc"
)

type StationMessageSeverity int

const (
	StationMessageSeverityNormal StationMessageSeverity = iota
	StationMessageSeverityMinor
	StationMessageSeverityMajor
	StationMessageSeveritySevere
)

var StationMessageStrings = map[StationMessageSeverity]string{
	StationMessageSeverityNormal: "Nominal",
	StationMessageSeverityMinor:  "Minor",
	StationMessageSeverityMajor:  "Major",
	StationMessageSeveritySevere: "Severe",
}

func (s StationMessageSeverity) String() string {
	if str, ok := StationMessageStrings[s]; ok {
		return str
	}
	return "Unknown Severity"
}
