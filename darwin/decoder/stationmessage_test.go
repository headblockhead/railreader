package decoder

import "testing"

var stationMessageTestCases = map[string]StationMessage{
	`<OW id="12345" cat="Misc" sev="1" suppress="true">
		<Station crs="LDS" />
		<Station crs="MAN" />
		<Msg><p>Line 0.</p><p>Line 1.</p><p>Line 2 and an <a href="http://example.com">example link</a> and some other text.</p></Msg>
	</OW>`: {
		ID:        "12345",
		Category:  "Misc",
		Severity:  1,
		Supressed: true,
		Stations: []StationCRS{
			{CRS: "LDS"},
			{CRS: "MAN"},
		},
		Message: XHTMLBody{
			Content: `<p>Line 0.</p><p>Line 1.</p><p>Line 2 and an <a href="http://example.com">example link</a> and some other text.</p>`,
		},
	},
}

func TestUnmarshalStationMessage(t *testing.T) {
	if err := TestUnmarshal(stationMessageTestCases); err != nil {
		t.Error(err)
	}
}
