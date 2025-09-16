package unmarshaller

import "testing"

var pportTestCases = []unmarshalTestCase[PushPortMessage]{
	{
		name: "timetable_files",
		xml: `
		<Pport ts="2025-08-23T08:00:00.12345+01:00" version="18.0">
			<TimeTableId ttfile="20250823080000_v8.xml.gz" ttreffile="20250823080000_ref_v99.xml.gz">20250823080000</TimeTableId>
		</Pport>
		`,
		expected: PushPortMessage{
			Timestamp: "2025-08-23T08:00:00.12345+01:00",
			Version:   "18.0",
			NewFiles: &NewFiles{
				TimetableFile: "20250823080000_v8.xml.gz",
				ReferenceFile: "20250823080000_ref_v99.xml.gz",
				Prefix:        "20250823080000",
			},
		},
	},
	{
		name: "status_update",
		xml: `
		<Pport ts="2025-08-23T08:00:00.12345+01:00" version="18.0">
			<FailureResp code="HBOK" requestSource="CIS1" requestID="01234567">Darwin Status Response</FailureResp>
		</Pport>
		`,
		expected: PushPortMessage{
			Timestamp: "2025-08-23T08:00:00.12345+01:00",
			Version:   "18.0",
			StatusUpdate: &Status{
				SourceSystem: pointerTo("CIS1"),
				RequestID:    pointerTo("01234567"),
				Code:         StatusCodeOK,
				Description:  "Darwin Status Response",
			},
		},
	},
	{
		name: "update_response",
		xml: `
		<Pport ts="2025-08-23T08:00:00.12345+01:00" version="18.0">
			<uR updateOrigin="CIS" requestSource="CIS1" requestID="01234567">
				<!-- data omitted -->
			</uR>
		</Pport>
		`,
		expected: PushPortMessage{
			Timestamp: "2025-08-23T08:00:00.12345+01:00",
			Version:   "18.0",
			UpdateResponse: &Response{
				Source:       pointerTo("CIS"),
				SourceSystem: pointerTo("CIS1"),
				RequestID:    pointerTo("01234567"),
			},
		},
	},
}

func TestUnmarshalPushPortMessage(t *testing.T) {
	testUnmarshal(t, pportTestCases)
}
