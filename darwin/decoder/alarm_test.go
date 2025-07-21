package decoder

import "github.com/headblockhead/railreader"

var feedFailure = ""

var AlarmTestCases = map[string]Alarm{
	`
	`: {
		ClearedAlarm: "01234",
		NewAlarm: NewAlarm{
			ID:                         "01234",
			TrainDescriptorAreaFailure: railreader.TrainDescriptorArea("Y2"),
			TrainDescriptorFeedFailure: &feedFailure,
			TyrellFeedFailure:          &feedFailure,
		},
	},
}
