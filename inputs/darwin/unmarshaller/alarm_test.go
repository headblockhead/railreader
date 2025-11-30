package unmarshaller

import (
	"testing"
)

var alarmTestCases = []unmarshalTestCase[Alarm]{
	{
		name: "all_fields_are_stored",
		xml: `
		<alarm>
			<clear>132</clear>
			<set id="325">
				<tdAreaFail>Y2</tdAreaFail>
				<tdFeedFail/>
				<tyrellFeedFail/>
			</set>
		</alarm>
		`,
		expected: Alarm{
			ClearedAlarm: pointerTo(132),
			NewAlarm: &NewAlarm{
				ID:                 325,
				TDFailure:          pointerTo("Y2"),
				TDTotalFailure:     true,
				TyrellTotalFailure: true,
			},
		},
	},
	{
		name: "feed_failures_are_false_when_not_given",
		xml: `
		<alarm>
			<set id="578">
				<tdAreaFail>Y1</tdAreaFail>
			</set>
		</alarm>
		`,
		expected: Alarm{
			NewAlarm: &NewAlarm{
				ID:        578,
				TDFailure: pointerTo("Y1"),
			},
		},
	},
	{
		name: "cleared_alarm_is_nil_when_not_set",
		xml: `
		<alarm>
			<set id="863"></set>
		</alarm>
		`,
		expected: Alarm{
			ClearedAlarm: nil,
			NewAlarm:     &NewAlarm{ID: 863},
		},
	},
	{
		name: "new_alarm_is_nil_when_not_set",
		xml: `
		<alarm>
			<clear>132</clear>
		</alarm>
		`,
		expected: Alarm{
			ClearedAlarm: pointerTo(132),
			NewAlarm:     nil,
		},
	},
}

func TestUnmarshalAlarm(t *testing.T) {
	testUnmarshal(t, alarmTestCases)
}
