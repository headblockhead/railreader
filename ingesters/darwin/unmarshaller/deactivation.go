package unmarshaller

// Deactivation is sent to indicate a RID is expected to receive no further updates, and shouldn't be displayed publicly.
// A deactivation can be un-done by a subsequent Schedule with the same RID.
type Deactivation struct {
	// RID is the unique 16-character ID for the specific train that has been deactivated.
	RID string `xml:"rid,attr"`
}
