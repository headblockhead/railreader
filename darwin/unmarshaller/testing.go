package unmarshaller

import (
	"bytes"
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
			reader := bytes.NewReader([]byte(tc.xml))
			decoder := xml.NewDecoder(reader)
			var actual T
			if err := decoder.Decode(&actual); err != nil {
				t.Fatalf("unexpected error during unmarshal: %v", err)
			}
			if !cmp.Equal(actual, tc.expected) {
				t.Fatalf("actual result and expected result differ: %s", cmp.Diff(actual, tc.expected))
			}
		})
	}
}
