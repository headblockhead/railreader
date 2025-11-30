package unmarshaller

import (
	"testing"

	"github.com/headblockhead/railreader"
)

var associationTestCases = []unmarshalTestCase[Association]{
	{
		name: "all_fields_are_stored",
		xml: `
		<association tiploc="ABCD" category="NP" isCancelled="true" isDeleted="true">
			<main rid="012345678901234" wta="00:01" wtd="00:02" wtp="00:03" pta="00:04" ptd="00:05"/>
			<assoc rid="012345678901235" wta="00:06" wtd="00:07" wtp="00:08" pta="00:09" ptd="00:10"/>
		</association>
		`,
		expected: Association{
			TIPLOC:    "ABCD",
			Category:  railreader.AssociationNext,
			Cancelled: true,
			Deleted:   true,
			MainService: AssociatedService{
				RID: "012345678901234",
				LocationTimeIdentifiers: LocationTimeIdentifiers{
					WorkingArrivalTime:   pointerTo("00:01"),
					WorkingDepartureTime: pointerTo("00:02"),
					WorkingPassingTime:   pointerTo("00:03"),
					PublicArrivalTime:    pointerTo("00:04"),
					PublicDepartureTime:  pointerTo("00:05"),
				},
			},
			AssociatedService: AssociatedService{
				RID: "012345678901235",
				LocationTimeIdentifiers: LocationTimeIdentifiers{
					WorkingArrivalTime:   pointerTo("00:06"),
					WorkingDepartureTime: pointerTo("00:07"),
					WorkingPassingTime:   pointerTo("00:08"),
					PublicArrivalTime:    pointerTo("00:09"),
					PublicDepartureTime:  pointerTo("00:10"),
				},
			},
		},
	},
	{
		name: "optional_fields_are_stored_correctly_when_not_set",
		xml: `
		<association tiploc="EFGH" category="NP">
			<main rid="012345678901236" wta="10:01"/>
			<assoc rid="012345678901237" wtd="10:02"/>
		</association>
		`,
		expected: Association{
			TIPLOC:    "EFGH",
			Category:  railreader.AssociationNext,
			Cancelled: false,
			Deleted:   false,
			MainService: AssociatedService{
				RID: "012345678901236",
				LocationTimeIdentifiers: LocationTimeIdentifiers{
					WorkingArrivalTime:   pointerTo("10:01"),
					WorkingDepartureTime: nil,
					WorkingPassingTime:   nil,
					PublicArrivalTime:    nil,
					PublicDepartureTime:  nil,
				},
			},
			AssociatedService: AssociatedService{
				RID: "012345678901237",
				LocationTimeIdentifiers: LocationTimeIdentifiers{
					WorkingArrivalTime:   nil,
					WorkingDepartureTime: pointerTo("10:02"),
					WorkingPassingTime:   nil,
					PublicArrivalTime:    nil,
					PublicDepartureTime:  nil,
				},
			},
		},
	},
}

func TestUnmarshalAssociation(t *testing.T) {
	testUnmarshal(t, associationTestCases)
}
