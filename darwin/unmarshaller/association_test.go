package unmarshaller

import (
	"testing"
)

var associationTestCases = map[string]Association{
	// Test that all fields are parsed correctly.
	`<association tiploc="ABCD" category="NP" isCancelled="true" isDeleted="true">
		<main rid="012345678901234" wta="00:01" wtd="00:02" wtp="00:03" pta="00:04" ptd="00:05"/>
		<assoc rid="012345678901235" wta="00:06" wtd="00:07" wtp="00:08" pta="00:09" ptd="00:10"/>
	</association>`: {
		TIPLOC:    "ABCD",
		Category:  "NP",
		Cancelled: true,
		Deleted:   true,
		MainService: AssociatedService{
			RID: "012345678901234",
			LocationTimeIdentifiers: LocationTimeIdentifiers{
				WorkingArrivalTime:   "00:01",
				WorkingDepartureTime: "00:02",
				WorkingPassingTime:   "00:03",
				PublicArrivalTime:    "00:04",
				PublicDepartureTime:  "00:05",
			},
		},
		AssociatedService: AssociatedService{
			RID: "012345678901235",
			LocationTimeIdentifiers: LocationTimeIdentifiers{
				WorkingArrivalTime:   "00:06",
				WorkingDepartureTime: "00:07",
				WorkingPassingTime:   "00:08",
				PublicArrivalTime:    "00:09",
				PublicDepartureTime:  "00:10",
			},
		},
	},
	// Test that unspceified fields are handled correctly.
	`<association tiploc="EFGH" category="NP">
		<main rid="012345678901236"/>
		<assoc rid="012345678901237"/>
	</association>`: {
		TIPLOC:   "EFGH",
		Category: "NP",
		MainService: AssociatedService{
			RID: "012345678901236",
		},
		AssociatedService: AssociatedService{
			RID: "012345678901237",
		},
	},
}

func TestUnmarshalAssociation(t *testing.T) {
	if err := TestUnmarshal(associationTestCases); err != nil {
		t.Error(err)
	}
}
