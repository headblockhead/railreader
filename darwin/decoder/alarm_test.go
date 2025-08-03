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
		ClearedAlarm: 132,
		NewAlarm: &NewAlarm{
			ID:                             325,
			TrainDescriberAreaFailure:      "Y2",
			TrainDescriberTotalFeedFailure: false,
			TyrellTotalFeedFailure:         false,
		},
	},
	// Test that the absence of a failure element is treated as false.
	`<alarm>
		<set id="329">
			<tyrellFeedFail/>
		</set>
	</alarm>`: {
		ClearedAlarm: 0,
		NewAlarm: &NewAlarm{
			ID:                             329,
			TrainDescriberAreaFailure:      "",
			TrainDescriberTotalFeedFailure: false,
			TyrellTotalFeedFailure:         true,
		},
	},
	// Test that the absence of a set element creates a nil pointer.
	`<alarm>
		<clear>132</clear>
	</alarm>`: {
		ClearedAlarm: 132,
		NewAlarm:     nil,
	},
}

func TestUnmarshalAlarm(t *testing.T) {
	if err := TestUnmarshal(alarmTestCases); err != nil {
		t.Error(err)
	}
}
