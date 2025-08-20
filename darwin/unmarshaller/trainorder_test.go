package unmarshaller

import "testing"

var trainOrderTestCases = map[string]TrainOrder{
	`<trainOrder tiploc="ABCD" crs="LDS" platform="17a">
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
	</trainOrder>`: {
		TIPLOC:   "ABCD",
		CRS:      "LDS",
		Platform: "17a",
		Services: TrainOrderServices{
			First: TrainOrderService{
				RIDAndTime: OrderedService{
					RID: "012345678901234",
					LocationTimeIdentifiers: LocationTimeIdentifiers{
						WorkingArrivalTime:   "00:01",
						WorkingDepartureTime: "00:02",
						WorkingPassingTime:   "00:03",
						PublicArrivalTime:    "00:04",
						PublicDepartureTime:  "00:05",
					},
				},
			},
			Second: TrainOrderService{
				Headcode: "2C04",
			},
			Third: &TrainOrderService{
				Headcode: "2C05",
			},
		},
	},
	`<trainOrder tiploc="EFGH" crs="MAN" platform="1">
		<clear/>
	</trainOrder>`: {
		TIPLOC:     "EFGH",
		CRS:        "MAN",
		Platform:   "1",
		ClearOrder: true,
	},
}

func TestUnmarshalTrainOrder(t *testing.T) {
	if err := TestUnmarshal(trainOrderTestCases); err != nil {
		t.Error(err)
	}
}
