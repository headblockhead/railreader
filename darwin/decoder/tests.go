package decoder

import (
	"encoding/xml"
	"errors"
	"fmt"

	"github.com/google/go-cmp/cmp"
)

func TestUnmarshal[T any](cases map[string]T) error {
	for inputXML, expectedOutput := range cases {
		var actualOutput T
		if err := xml.Unmarshal([]byte(inputXML), &actualOutput); err != nil {
			return fmt.Errorf("failed to unmarshal input %q: %w", inputXML, err)
		}
		if !cmp.Equal(expectedOutput, actualOutput) {
			return errors.New(cmp.Diff(expectedOutput, actualOutput))
		}
	}
	return nil
}
