package decoder

import "github.com/headblockhead/railreader"

type HeadcodeChange struct {
	OldHeadcode string `xml:"incorrectTrainID"`
	NewHeadcode string `xml:"correctTrainID"`

	// TrainDescriberLocation contains some Train Describer location information.
	TrainDescriberLocation TrainDescriberLocation `xml:"berth"`
}

type TrainDescriberLocation struct {
	// Area is a two-character code representing the TD area the train is in.
	Area railreader.TrainDescriberArea `xml:"area,attr"`
	// Berth is the four-character Train Describer berth (area of track dictated usually by a signal) where the train is (likely) located.
	Berth string `xml:",chardata"`
}
