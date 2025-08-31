package unmarshaller

import "testing"

var trainOrderTestCases = []unmarshalTestCase[TrainOrder]{
	{
		name: "all_fields_are_stored_in_a_service_set",
		xml: `
		<trainOrder tiploc="ABCD" crs="LDS" platform="17a">
			<set>
				<first>
					<rid wta="00:01" wtd="00:02" wtp="00:03" pta="00:04" ptd="00:05">012345678901234</rid>
				</first>
				<second>
					<trainID>2C04</trainID>
				</second>
				<third>
					<trainID>2C05</trainID>
				</third>
			</set>
		</trainOrder>
		`,
		expected: TrainOrder{
			TIPLOC:   "ABCD",
			CRS:      "LDS",
			Platform: "17a",
			Services: &TrainOrderServices{
				First: TrainOrderService{
					RIDAndTime: &OrderedService{
						RID: "012345678901234",
						LocationTimeIdentifiers: LocationTimeIdentifiers{
							WorkingArrivalTime:   pointerTo(TrainTime("00:01")),
							WorkingDepartureTime: pointerTo(TrainTime("00:02")),
							WorkingPassingTime:   pointerTo(TrainTime("00:03")),
							PublicArrivalTime:    pointerTo(TrainTime("00:04")),
							PublicDepartureTime:  pointerTo(TrainTime("00:05")),
						},
					},
				},
				Second: TrainOrderService{
					Headcode: pointerTo("2C04"),
				},
				Third: &TrainOrderService{
					Headcode: pointerTo("2C05"),
				},
			},
		},
	},
	{
		name: "trainorder_can_be_cleared",
		xml: `<trainOrder tiploc="EFGH" crs="MAN" platform="1">
		<clear/>
	</trainOrder>`,
		expected: TrainOrder{
			TIPLOC:     "EFGH",
			CRS:        "MAN",
			Platform:   "1",
			ClearOrder: true,
		},
	},
}

func TestUnmarshalTrainOrder(t *testing.T) {
	testUnmarshal(t, trainOrderTestCases)
}
