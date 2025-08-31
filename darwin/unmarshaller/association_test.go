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
					WorkingArrivalTime:   pointerTo(TrainTime("00:01")),
					WorkingDepartureTime: pointerTo(TrainTime("00:02")),
					WorkingPassingTime:   pointerTo(TrainTime("00:03")),
					PublicArrivalTime:    pointerTo(TrainTime("00:04")),
					PublicDepartureTime:  pointerTo(TrainTime("00:05")),
				},
			},
			AssociatedService: AssociatedService{
				RID: "012345678901235",
				LocationTimeIdentifiers: LocationTimeIdentifiers{
					WorkingArrivalTime:   pointerTo(TrainTime("00:06")),
					WorkingDepartureTime: pointerTo(TrainTime("00:07")),
					WorkingPassingTime:   pointerTo(TrainTime("00:08")),
					PublicArrivalTime:    pointerTo(TrainTime("00:09")),
					PublicDepartureTime:  pointerTo(TrainTime("00:10")),
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
					WorkingArrivalTime:   pointerTo(TrainTime("10:01")),
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
					WorkingDepartureTime: pointerTo(TrainTime("10:02")),
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
