package decoder

type HeadcodeChange struct {
	OldHeadcode string `xml:"incorrectTrainID"`
	NewHeadcode string `xml:"correctTrainID"`

	// TDLocation contains some Train Describer location information.
	TDLocation TDLocation `xml:"berth"`
}

type TDLocation struct {
	// Area is a two-character code representing the TD area the train is in.
	Area string `xml:"area,attr"`
	// Berth is the four-character Train Descriptor berth (area of track dictated by a signal) where the train is (likely) located.
	Berth string `xml:",chardata"`
}
