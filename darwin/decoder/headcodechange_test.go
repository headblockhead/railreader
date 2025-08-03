package decoder

import "testing"

var headcodeChangeTestCases = map[string]HeadcodeChange{
	`<trackingID>
		<incorrectTrainID>2C04</incorrectTrainID>
		<correctTrainID>2C05</correctTrainID>
		<berth area="Y2">L3608</berth>
	</trackingID>`: {
		OldHeadcode: "2C04",
		NewHeadcode: "2C05",
		TrainDescriberLocation: TrainDescriberLocation{
			Area:  "Y2",
			Berth: "L3608",
		},
	},
}

func TestUnmarshalHeadcodeChange(t *testing.T) {
	if err := TestUnmarshal(headcodeChangeTestCases); err != nil {
		t.Error(err)
	}
}
