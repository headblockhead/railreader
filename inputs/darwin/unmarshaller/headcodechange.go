package unmarshaller

type HeadcodeChange struct {
	OldHeadcode string `xml:"incorrectTrainID"`
	NewHeadcode string `xml:"correctTrainID"`

	// TDLocation contains some Train Describer location information.
	TDLocation TDLocation `xml:"berth"`
}

type TDLocation struct {
	Describer string `xml:"area,attr"`
	// Berth where the train is (likely) located.
	Berth string `xml:",chardata"`
}
