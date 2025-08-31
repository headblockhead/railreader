package unmarshaller

import (
	"testing"

	"github.com/headblockhead/railreader"
)

var scheduleTestCases = []unmarshalTestCase[Schedule]{
	{
		name: "all_fields_are_stored",
		xml: `
		<schedule rid="012345678901234" uid="A00001" trainId="2C04" rsid="GW123456" ssd="2025-08-23" toc="GW" status="B" trainCat="BR" isPassengerSvc="true" isActive="true" deleted="true" isCharter="true">
			<OR tpl="MNOP" act="TBT " planAct="TB" can="true" fid="012345678901234-001" affectedByDiversion="true" pta="00:01" ptd="00:02" wta="00:03" wtd="00:04" fd="QRST">
				<cancelReason tiploc="UVWX" near="true">102</cancelReason>
			</OR>
			<OPOR wta="00:05" wtd="00:06" />
			<IP pta="00:07" ptd="00:08" wta="00:09" wtd="00:10" rdelay="2" fd="YZAB"/>
			<OPIP wta="00:11" wtd="00:12" rdelay="3" />
			<PP wtp="00:13" rdelay="4" />
			<DT pta="00:14" ptd="00:15" wta="00:16" wtd="00:17" rdelay="5"/>
			<OPDT wta="00:18" wtd="00:19" rdelay="6"/>
			<cancelReason tiploc="ABCD" near="true">100</cancelReason>
			<divertedVia>EFGH</divertedVia>
			<diversionReason tiploc="IJKL" near="true">101</diversionReason>
		</schedule>
		`,
		expected: Schedule{
			TrainIdentifiers: TrainIdentifiers{
				RID:                "012345678901234",
				UID:                "A00001",
				ScheduledStartDate: "2025-08-23",
			},
			Headcode:         "2C04",
			RetailServiceID:  pointerTo("GW123456"),
			TOC:              "GW",
			Service:          railreader.ServiceBus,
			Category:         railreader.CategoryBusReplacement,
			PassengerService: true,
			Active:           true,
			Deleted:          true,
			Charter:          true,
			CancellationReason: &DisruptionReason{
				TIPLOC:   pointerTo(railreader.TimingPointLocationCode("ABCD")),
				Near:     true,
				ReasonID: 100,
			},
			DivertedVia: pointerTo(railreader.TimingPointLocationCode("EFGH")),
			DiversionReason: &DisruptionReason{
				TIPLOC:   pointerTo(railreader.TimingPointLocationCode("IJKL")),
				Near:     true,
				ReasonID: 101,
			},
			Locations: []ScheduleLocation{
				{
					Type: LocationTypeOrigin,
					Origin: &OriginLocation{
						LocationBase: LocationBase{
							TIPLOC:              "MNOP",
							Activities:          pointerTo("TBT "),
							PlannedActivities:   pointerTo("TB"),
							Cancelled:           true,
							FormationID:         pointerTo("012345678901234-001"),
							AffectedByDiversion: true,
							CancellationReason: &DisruptionReason{
								TIPLOC:   pointerTo(railreader.TimingPointLocationCode("UVWX")),
								Near:     true,
								ReasonID: 102,
							},
						},
						PublicArrivalTime:    pointerTo(TrainTime("00:01")),
						PublicDepartureTime:  pointerTo(TrainTime("00:02")),
						WorkingArrivalTime:   pointerTo(TrainTime("00:03")),
						WorkingDepartureTime: TrainTime("00:04"),
						FalseDestination:     pointerTo(railreader.TimingPointLocationCode("QRST")),
					},
				},
				{
					Type: LocationTypeOperationalOrigin,
					OperationalOrigin: &OperationalOriginLocation{
						WorkingArrivalTime:   pointerTo(TrainTime("00:05")),
						WorkingDepartureTime: TrainTime("00:06"),
					},
				},
				{
					Type: LocationTypeIntermediate,
					Intermediate: &IntermediateLocation{
						PublicArrivalTime:    pointerTo(TrainTime("00:07")),
						PublicDepartureTime:  pointerTo(TrainTime("00:08")),
						WorkingArrivalTime:   TrainTime("00:09"),
						WorkingDepartureTime: TrainTime("00:10"),
						RoutingDelay:         pointerTo(2),
						FalseDestination:     pointerTo(railreader.TimingPointLocationCode("YZAB")),
					},
				},
				{
					Type: LocationTypeOperationalIntermediate,
					OperationalIntermediate: &OperationalIntermediateLocation{
						WorkingArrivalTime:   TrainTime("00:11"),
						WorkingDepartureTime: TrainTime("00:12"),
						RoutingDelay:         pointerTo(3),
					},
				},
				{
					Type: LocationTypeIntermediatePassing,
					IntermediatePassing: &IntermediatePassingLocation{
						WorkingPassingTime: TrainTime("00:13"),
						RoutingDelay:       pointerTo(4),
					},
				},
				{
					Type: LocationTypeDestination,
					Destination: &DestinationLocation{
						PublicArrivalTime:    pointerTo(TrainTime("00:14")),
						PublicDepartureTime:  pointerTo(TrainTime("00:15")),
						WorkingArrivalTime:   TrainTime("00:16"),
						WorkingDepartureTime: pointerTo(TrainTime("00:17")),
						RoutingDelay:         pointerTo(5),
					},
				},
				{
					Type: LocationTypeOperationalDestination,
					OperationalDestination: &OperationalDestinationLocation{
						WorkingArrivalTime:   TrainTime("00:18"),
						WorkingDepartureTime: pointerTo(TrainTime("00:19")),
						RoutingDelay:         pointerTo(6),
					},
				},
			},
		},
	},
	{
		name: "default_values",
		xml:  `<schedule/>`,
		expected: Schedule{
			Service:          railreader.ServicePassengerOrParcelTrain,
			Category:         railreader.CategoryPassenger,
			PassengerService: true,
			Active:           true,
		},
	},
}

func TestUnmarshalSchedule(t *testing.T) {
	testUnmarshal(t, scheduleTestCases)
}
