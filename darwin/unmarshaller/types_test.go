package unmarshaller

import (
	"encoding/xml"
	"testing"
)

type trueIfPresentTester struct {
	XMLName xml.Name      `xml:"container"`
	Example TrueIfPresent `xml:"example"`
}

var trueIfPresentTestCases = []unmarshalTestCase[trueIfPresentTester]{
	{
		name: "present",
		xml: `
		<container>
			<example/>
		</container>
		`,
		expected: trueIfPresentTester{
			XMLName: xml.Name{Local: "container"},
			Example: true,
		},
	},
	{
		name: "absent",
		xml: `
		<container>
		</container>
		`,
		expected: trueIfPresentTester{
			XMLName: xml.Name{Local: "container"},
			Example: false,
		},
	},
}

func TestUnmarshalTrueIfPresent(t *testing.T) {
	testUnmarshal(t, trueIfPresentTestCases)
}
