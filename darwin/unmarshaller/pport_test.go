package unmarshaller

import "testing"

var pportTestCases = map[string]PushPortMessage{
	`<Pport ts="2025-08-23T08:00:00.12345+01:00" version="18.0">
		<TimeTableId ttfile="20250823080000_v8.xml.gz" ttreffile="20250823080000_ref_v99.xml.gz">20250823080000</TimeTableId>
	</Pport>`: {
		Timestamp: "2025-08-23T08:00:00.12345+01:00",
		Version:   "18.0",
		NewTimetableFiles: &TimetableFiles{
			TimetableFile:          "20250823080000_v8.xml.gz",
			TimetableReferenceFile: "20250823080000_ref_v99.xml.gz",
			TimeTableId:            "20250823080000",
		},
	},
	`<Pport ts="2025-08-23T08:00:00.12345+01:00" version="18.0">
		<FailureResp code="HBOK" requestSource="CIS1" requestID="01234567">Darwin Status Response</FailureResp>
	</Pport>`: {
		Timestamp: "2025-08-23T08:00:00.12345+01:00",
		Version:   "18.0",
		StatusUpdate: &Status{
			SourceSystem: "CIS1",
			RequestID:    "01234567",
			Code:         StatusCodeOK,
			Description:  "Darwin Status Response",
		},
	},
	`<Pport ts="2025-08-23T08:00:00.12345+01:00" version="18.0">
		<uR updateOrigin="CIS" requestSource="CIS1" requestID="01234567">
			<!-- data omitted -->
		</uR>
	</Pport>`: {
		Timestamp: "2025-08-23T08:00:00.12345+01:00",
		Version:   "18.0",
		UpdateResponse: &Response{
			Source:       "CIS",
			SourceSystem: "CIS1",
			RequestID:    "01234567",
		},
	},
}

func TestUnmarshalPushPortMessage(t *testing.T) {
	if err := TestUnmarshal(pportTestCases); err != nil {
		t.Error(err)
	}
}
