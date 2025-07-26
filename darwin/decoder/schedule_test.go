package decoder

import (
	"testing"

	"github.com/headblockhead/railreader"
)

var scheduleTestCases = map[string]Schedule{
	`<schedule rid="012345678901234" uid="A00001" trainId="2C04" rsid="GW123456" ssd="2006-01-02" toc="GW" status="B" trainCat="BR" isPassengerSvc="true" isActive="true" deleted="true" isCharter="true">
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
	</schedule>`: {
		TrainIdentifiers: TrainIdentifiers{
			RID:                "012345678901234",
			UID:                "A00001",
			ScheduledStartDate: "2006-01-02",
		},
		Headcode:              "2C04",
		RetailServiceID:       "GW123456",
		TrainOperatingCompany: "GW",
		Service:               railreader.ServiceBus,
		Category:              railreader.CategoryBusReplacement,
		PassengerService:      true,
		Active:                true,
		Deleted:               true,
		Charter:               true,
		CancellationReason: &DisruptionReason{
			TIPLOC: "ABCD",
			Near:   true,
			Reason: "100",
		},
		DivertedVia: "EFGH",
		DiversionReason: &DisruptionReason{
			TIPLOC: "IJKL",
			Near:   true,
			Reason: "101",
		},
		Locations: []LocationGeneric{
			{
				Type: LocationTypeOrigin,
				OriginLocation: &OriginLocation{
					LocationSchedule: LocationSchedule{
						TIPLOC:              "MNOP",
						Activities:          "TBT ",
						PlannedActivities:   "TB",
						Cancelled:           true,
						FormationID:         "012345678901234-001",
						AffectedByDiversion: true,
						CancellationReason: &DisruptionReason{
							TIPLOC: "UVWX",
							Near:   true,
							Reason: "102",
						},
					},
					PublicArrivalTime:    "00:01",
					PublicDepartureTime:  "00:02",
					WorkingArrivalTime:   "00:03",
					WorkingDepartureTime: "00:04",
					FalseDestination:     "QRST",
				},
			},
			{
				Type: LocationTypeOperationalOrigin,
				OperationalOriginLocation: &OperationalOriginLocation{
					WorkingArrivalTime:   "00:05",
					WorkingDepartureTime: "00:06",
				},
			},
			{
				Type: LocationTypeIntermediate,
				IntermediateLocation: &IntermediateLocation{
					PublicArrivalTime:    "00:07",
					PublicDepartureTime:  "00:08",
					WorkingArrivalTime:   "00:09",
					WorkingDepartureTime: "00:10",
					RoutingDelay:         2,
					FalseDestination:     "YZAB",
				},
			},
			{
				Type: LocationTypeOperationalIntermediate,
				OperationalIntermediateLocation: &OperationalIntermediateLocation{
					WorkingArrivalTime:   "00:11",
					WorkingDepartureTime: "00:12",
					RoutingDelay:         3,
				},
			},
			{
				Type: LocationTypeIntermediatePassing,
				IntermediatePassingLocation: &IntermediatePassingLocation{
					WorkingPassingTime: "00:13",
					RoutingDelay:       4,
				},
			},
			{
				Type: LocationTypeDestination,
				DestinationLocation: &DestinationLocation{
					PublicArrivalTime:    "00:14",
					PublicDepartureTime:  "00:15",
					WorkingArrivalTime:   "00:16",
					WorkingDepartureTime: "00:17",
					RoutingDelay:         5,
				},
			},
			{
				Type: LocationTypeOperationalDestination,
				OperationalDestinationLocation: &OperationalDestinationLocation{
					WorkingArrivalTime:   "00:18",
					WorkingDepartureTime: "00:19",
					RoutingDelay:         6,
				},
			},
		},
	},
	// Test default values
	`<schedule/>`: {
		Service:          railreader.ServicePassengerOrParcelTrain,
		Category:         railreader.CategoryPassenger,
		PassengerService: true,
		Active:           true,
	},
}

func TestUnmarshalSchedule(t *testing.T) {
	if err := TestUnmarshal(scheduleTestCases); err != nil {
		t.Error(err)
	}
}
