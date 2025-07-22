package decoder

import "testing"

// TODO
var pportTestCases = map[string]PushPortMessage{
	`<Pport ts="2006-01-02T15:04:05.999999999-07:00" version="18.0">
	</Pport>`: {},
}

func TestUnmarshalPushPortMessage(t *testing.T) {
	if err := TestUnmarshal(pportTestCases); err != nil {
		t.Errorf("%v", err)
	}
}
