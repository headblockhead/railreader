package decoder

type HeadcodeChange struct {
	OldHeadcode string `xml:"incorrectTrainID"`
	NewHeadcode string `xml:"correctTrainID"`

	// TDLocation contains some Train Descriptor location information.
	TDLocation struct {
		// Berth is the Train Descriptor berth (area of track/signal) where the train is (likely) located.
		Berth string `xml:",chardata"`
		// Area is a two-character code representing the TD area the train is in.
		Area string `xml:"area,attr"`
	} `xml:"berth"`
}
