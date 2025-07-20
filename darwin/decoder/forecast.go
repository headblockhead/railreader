package decoder

type ActualAndForecastInformation struct {
	// RID is the unique 16-character ID for a specific train.
	RID string `xml:"rid,attr"`
	// UID is (despite the name) a non-unique 6-character ID for a route at a time of day.
	UID string `xml:"uid,attr"`
	// ScheduledStartDate is in YYYY-MM-DD format.
	ScheduledStartDate string `xml:"ssd,attr"`
	// ReverseFormation indicates whether a train that divides will run in reverse formation after the dividing location.
	ReverseFormation string `xml:"isReverseFormation,attr"`

	LateReason DisruptionReason   `xml:"LateReason"`
	Locations  []LocationForecast `xml:"Location"`
}
