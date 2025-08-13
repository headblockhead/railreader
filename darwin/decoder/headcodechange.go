package decoder

import "github.com/headblockhead/railreader"

type HeadcodeChange struct {
	OldHeadcode string `xml:"incorrectTrainID"`
	NewHeadcode string `xml:"correctTrainID"`

	// TrainDescriberLocation contains some Train Describer location information.
	TrainDescriberLocation TrainDescriberLocation `xml:"berth"`
}

type TrainDescriberLocation struct {
	Describer railreader.TrainDescriber `xml:"area,attr"`
	// Berth where the train is (likely) located.
	Berth railreader.TrainDescriberBerth `xml:",chardata"`
}
