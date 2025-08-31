package unmarshaller

import "testing"

var deactivationTestCases = []unmarshalTestCase[Deactivation]{
	{
		name: "all_fields_are_stored",
		xml: `
		<deactivated rid="012345678901234" />
		`,
		expected: Deactivation{
			RID: "012345678901234",
		},
	},
}

func TestUnmarshalDeactivation(t *testing.T) {
	testUnmarshal(t, deactivationTestCases)
}
