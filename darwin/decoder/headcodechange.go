package decoder

import "github.com/headblockhead/railreader"

type HeadcodeChange struct {
	OldHeadcode string `xml:"incorrectTrainID"`
	NewHeadcode string `xml:"correctTrainID"`

	// TDLocation contains some Train Describer location information.
	TDLocation TDLocation `xml:"berth"`
}

type TDLocation struct {
	Describer railreader.TrainDescriber `xml:"area,attr"`
	// Berth where the train is (likely) located.
	Berth railreader.TrainDescriberBerth `xml:",chardata"`
}
