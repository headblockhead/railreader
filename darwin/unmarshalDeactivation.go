package darwin

// DeactivationInformation is sent to indicate a RID is expected to recieve no further updates, and shouldn't be displayed publicly.
// A deactivation can be un-done by a subsequent ScheduleInformation with the same RID.
type DeactivationInformation struct {
	// RID is the unique 16-character ID for the specific train+schedule+time combo that has been deactivated.
	RID string `xml:"rid,attr"`
}
