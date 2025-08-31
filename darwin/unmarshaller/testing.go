package unmarshaller

import (
	"encoding/xml"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type unmarshalTestCase[T any] struct {
	name     string
	xml      string
	expected T
}

func testUnmarshal[T any](t *testing.T, cases []unmarshalTestCase[T]) {
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var actual T
			if err := xml.Unmarshal([]byte(tc.xml), &actual); err != nil {
				t.Fatalf("failed to unmarshal case XML: %v", err)
			}
			if !cmp.Equal(tc.expected, actual) {
				t.Fatalf("actual result and expected result differ: %s", cmp.Diff(actual, tc.expected))
			}
		})
	}
}
