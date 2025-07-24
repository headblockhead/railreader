package decoder

import "testing"

var deactivationTestCases = map[string]Deactivation{
	`<deactivated rid="012345678901234" />`: {
		RID: "012345678901234",
	},
}

func TestUnmarshalDeactivation(t *testing.T) {
	if err := TestUnmarshal(deactivationTestCases); err != nil {
		t.Error(err)
	}
}
