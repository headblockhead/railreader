package decoder

import (
	"testing"
)

var alarmTestCases = map[string]Alarm{
	// Test that main string-based elements can be decoded.
	`<alarm>
		<clear>132</clear>
		<set id="325">
			<tdAreaFail>Y2</tdAreaFail>
		</set>
	</alarm>`: {
		ClearedAlarm: "132",
		NewAlarm: &NewAlarm{
			ID:                         "325",
			TrainDescriptorAreaFailure: "Y2",
			TrainDescriptorFeedFailure: false,
			TyrellFeedFailure:          false,
		},
	},
	// Test that the absence of a failure element is treated as false.
	`<alarm>
		<set id="329">
			<tyrellFeedFail/>
		</set>
	</alarm>`: {
		ClearedAlarm: "",
		NewAlarm: &NewAlarm{
			ID:                         "329",
			TrainDescriptorAreaFailure: "",
			TrainDescriptorFeedFailure: false,
			TyrellFeedFailure:          true,
		},
	},
	// Test that the absence of a set element creates a nil pointer.
	`<alarm>
		<clear>132</clear>
	</alarm>`: {
		ClearedAlarm: "132",
		NewAlarm:     nil,
	},
}

func TestUnmarshalAlarm(t *testing.T) {
	if err := TestUnmarshal(alarmTestCases); err != nil {
		t.Error(err)
	}
}
