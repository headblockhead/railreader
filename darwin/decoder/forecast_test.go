package decoder

import (
	"testing"

	"github.com/headblockhead/railreader"
)

var forecastTestCases = map[string]ForecastTime{
	// Test decoding every possible field.
	`<TS rid="012345678901234" uid="A00001" ssd="2006-01-02" isReverseFormation="true">
		<LateReason tiploc="EFGH" near="true">100</LateReason>
	  <Location tpl="ABCD" wta="00:01" wtd="00:02" wtp="00:03" pta="00:04" ptd="00:05" >
			<arr et="00:12" wet="00:13" at="00:14" atRemoved="true" atClass="Manual" etmin="00:11" etUnknown="true" delayed="true" src="TRUST" srcInst="Auto"/>
			<dep et="00:16" wet="00:17" at="00:18" atRemoved="true" atClass="Manual" etmin="00:15" etUnknown="true" delayed="true" src="TRUST" srcInst="Auto"/>
			<pass et="00:20" wet="00:21" at="00:22" atRemoved="true" atClass="Manual" etmin="00:19" etUnknown="true" delayed="true" src="TRUST" srcInst="Auto"/>
		  <plat platsup="true" cisPlatsup="true" platsrc="M" conf="true">2</plat>
			<suppr>true</suppr>
			<length>3</length>
			<detachFront>true</detachFront>
			<LateReason tiploc="IJKL" near="true">101</LateReason>
			<!--I have not seen any examples of the status value yet, the use of "delay" here is a guess-->
			<uncertainty status="delay"><reason tiploc="MNOP" near="true">102</reason></uncertainty>
			<affectedBy>123456</affectedBy>
	  </Location>
	</TS>`: {
		RID:                "012345678901234",
		UID:                "A00001",
		ScheduledStartDate: "2006-01-02",
		ReverseFormation:   true,
		LateReason: &DisruptionReason{
			TIPLOC: "EFGH",
			Near:   true,
			Reason: "100",
		},
		Locations: []LocationForecast{
			{
				TIPLOC:               railreader.TIPLOC("ABCD"),
				WorkingArrivalTime:   "00:01",
				WorkingDepartureTime: "00:02",
				WorkingPassingTime:   "00:03",
				PublicArrivalTime:    "00:04",
				PublicDepartureTime:  "00:05",
				ArrivalData: &LocationForecastTimeData{
					EstimatedTime:        "00:12",
					WorkingTime:          "00:13",
					ActualTime:           "00:14",
					ActualTimeRevoked:    true,
					ActualTimeSource:     "Manual",
					EstimatedTimeMinimum: "00:11",
					EstimatedTimeUnknown: true,
					Delayed:              true,
					Source:               "TRUST",
					SourceSystem:         "Auto",
				},
				DepartureData: &LocationForecastTimeData{
					EstimatedTime:        "00:16",
					WorkingTime:          "00:17",
					ActualTime:           "00:18",
					ActualTimeRevoked:    true,
					ActualTimeSource:     "Manual",
					EstimatedTimeMinimum: "00:15",
					EstimatedTimeUnknown: true,
					Delayed:              true,
					Source:               "TRUST",
					SourceSystem:         "Auto",
				},
				PassingData: &LocationForecastTimeData{
					EstimatedTime:        "00:20",
					WorkingTime:          "00:21",
					ActualTime:           "00:22",
					ActualTimeRevoked:    true,
					ActualTimeSource:     "Manual",
					EstimatedTimeMinimum: "00:19",
					EstimatedTimeUnknown: true,
					Delayed:              true,
					Source:               "TRUST",
					SourceSystem:         "Auto",
				},
				Suppressed:        true,
				DetachesFromFront: true,
				LateReason: &DisruptionReason{
					TIPLOC: "IJKL",
					Near:   true,
					Reason: "101",
				},
				Uncertainty: &Uncertainty{
					Status: "delay",
					Reason: &DisruptionReason{
						TIPLOC: "MNOP",
						Near:   true,
						Reason: "102",
					},
				},
				AffectedBy: "123456",
				Length:     3,
				PlatformData: &PlatformData{
					Suppressed:      true,
					SuppressedByCIS: true,
					Source:          PlatformDataSourceManual,
					Confirmed:       true,
				},
			},
		},
	},
}

func TestUnmarshalForecast(t *testing.T) {
	if err := TestUnmarshal(forecastTestCases); err != nil {
		t.Error(err)
	}
}
