package unmarshaller

import "testing"

var headcodeChangeTestCases = []unmarshalTestCase[HeadcodeChange]{
	{
		name: "all_fields_are_stored",
		xml: `
		<trackingID>
			<incorrectTrainID>2C04</incorrectTrainID>
			<correctTrainID>2C05</correctTrainID>
			<berth area="Y2">L3608</berth>
		</trackingID>
		`,
		expected: HeadcodeChange{
			OldHeadcode: "2C04",
			NewHeadcode: "2C05",
			TDLocation: TDLocation{
				Describer: "Y2",
				Berth:     "L3608",
			},
		},
	},
}

func TestUnmarshalHeadcodeChange(t *testing.T) {
	testUnmarshal(t, headcodeChangeTestCases)
}
