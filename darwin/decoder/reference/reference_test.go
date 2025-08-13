package reference

import (
	"testing"

	"github.com/headblockhead/railreader/darwin/decoder"
)

var referenceTestCases = map[string]Reference{
	`<PportTimetableRef timetableId="20050121105940">
		<LocationRef tpl="MNCRPIC" crs="MAN" toc="RT" name="Manchester Piccadilly"/>
		<TocRef toc="RT" name="Network Rail" url="https://example.com" />
		<LateRunningReasons>
			<Reason code="1" reasontext="This train is delayed due to a problem"/>
		</LateRunningReasons>
		<CancellationReasons>
			<Reason code="1" reasontext="This train has been cancelled due to a problem"/>
		</CancellationReasons>
		<CISSource code="TH01" name="Southeastern"/>
</PportTimetableRef>`: {
		Locations: []Location{
			{
				Location: "MNCRPIC",
				CRS:      "MAN",
				TOC:      "RT",
				Name:     "Manchester Piccadilly",
			},
		},
		TrainOperatingCompanies: []TrainOperatingCompany{
			{
				ID:   "RT",
				Name: "Network Rail",
				URL:  "https://example.com",
			},
		},
		LateReasons: []Reason{
			{
				ReasonID:    1,
				Description: "This train is delayed due to a problem",
			},
		},
		CancellationReasons: []Reason{
			{
				ReasonID:    1,
				Description: "This train has been cancelled due to a problem",
			},
		},
		CustomerInformationSystemSources: []CustomerInformationSystemSource{
			{
				CIS:  "TH01",
				Name: "Southeastern",
			},
		},
	},
}

func TestUnmarshalReference(t *testing.T) {
	if err := decoder.TestUnmarshal(referenceTestCases); err != nil {
		t.Error(err)
	}
}
