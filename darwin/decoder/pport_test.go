package decoder

import "testing"

var pportTestCases = map[string]PushPortMessage{
	`<Pport ts="2006-01-02T15:04:05.999999999-07:00" version="18.0">
	<TimeTableId ttfile="20060102150405_v8.xml.gz" ttreffile="20060102150405_ref_v99.xml.gz">20060102150405</TimeTableId>
	</Pport>`: {
		Timestamp: "2006-01-02T15:04:05.999999999-07:00",
		Version:   "18.0",
		TimeTableID: &TimeTableId{
			TTfile:      "20060102150405_v8.xml.gz",
			TTRefFile:   "20060102150405_ref_v99.xml.gz",
			TimeTableId: "20060102150405",
		},
	},
	`<Pport ts="2006-01-02T15:04:05.999999999-07:00" version="18.0">
	<FailureResp code="HBOK" requestSource="CIS1" requestID="01234567">Darwin Status Response</FailureResp>
	</Pport>`: {
		Timestamp: "2006-01-02T15:04:05.999999999-07:00",
		Version:   "18.0",
		StatusResponse: &Status{
			RequestSourceSystem: "CIS1",
			RequestID:           "01234567",
			Code:                StatusCodeOK,
			Description:         "Darwin Status Response",
		},
	},
	`<Pport ts="2006-01-02T15:04:05.999999999-07:00" version="18.0">
	<uR updateOrigin="CIS" requestSource="CIS1" requestID="01234567">
	<!-- data omitted -->
	</uR>
	</Pport>`: {
		Timestamp: "2006-01-02T15:04:05.999999999-07:00",
		Version:   "18.0",
		UpdateResponse: &Response{
			UpdateOrigin:        "CIS",
			RequestSourceSystem: "CIS1",
			RequestID:           "01234567",
		},
	},
}

func TestUnmarshalPushPortMessage(t *testing.T) {
	if err := TestUnmarshal(pportTestCases); err != nil {
		t.Errorf("%v", err)
	}
}
