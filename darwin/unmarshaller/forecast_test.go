package unmarshaller

import (
	"testing"
)

var forecastTestCases = []unmarshalTestCase[ForecastTime]{
	{
		name: "all_fields_are_stored",
		xml: `
		<TS rid="012345678901234" uid="A00001" ssd="2025-08-23" isReverseFormation="true">
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
		</TS>
		`,
		expected: ForecastTime{
			TrainIdentifiers: TrainIdentifiers{
				RID:                "012345678901234",
				UID:                "A00001",
				ScheduledStartDate: "2025-08-23",
			},
			ReverseFormation: true,
			LateReason: &DisruptionReason{
				TIPLOC:   pointerTo(string("EFGH")),
				Near:     true,
				ReasonID: 100,
			},
			Locations: []ForecastLocation{
				{
					TIPLOC: "ABCD",
					LocationTimeIdentifiers: LocationTimeIdentifiers{
						WorkingArrivalTime:   pointerTo("00:01"),
						WorkingDepartureTime: pointerTo("00:02"),
						WorkingPassingTime:   pointerTo("00:03"),
						PublicArrivalTime:    pointerTo("00:04"),
						PublicDepartureTime:  pointerTo("00:05"),
					},
					ArrivalData: &ForecastTimes{
						EstimatedTime:        pointerTo("00:12"),
						WorkingTime:          pointerTo("00:13"),
						ActualTime:           pointerTo("00:14"),
						ActualTimeRevoked:    true,
						ActualTimeSource:     pointerTo("Manual"),
						EstimatedTimeMinimum: pointerTo("00:11"),
						EstimatedTimeUnknown: true,
						Delayed:              true,
						Source:               pointerTo("TRUST"),
						SourceSystem:         pointerTo("Auto"),
					},
					DepartureData: &ForecastTimes{
						EstimatedTime:        pointerTo("00:16"),
						WorkingTime:          pointerTo("00:17"),
						ActualTime:           pointerTo("00:18"),
						ActualTimeRevoked:    true,
						ActualTimeSource:     pointerTo("Manual"),
						EstimatedTimeMinimum: pointerTo("00:15"),
						EstimatedTimeUnknown: true,
						Delayed:              true,
						Source:               pointerTo("TRUST"),
						SourceSystem:         pointerTo("Auto"),
					},
					PassingData: &ForecastTimes{
						EstimatedTime:        pointerTo("00:20"),
						WorkingTime:          pointerTo("00:21"),
						ActualTime:           pointerTo("00:22"),
						ActualTimeRevoked:    true,
						ActualTimeSource:     pointerTo("Manual"),
						EstimatedTimeMinimum: pointerTo("00:19"),
						EstimatedTimeUnknown: true,
						Delayed:              true,
						Source:               pointerTo("TRUST"),
						SourceSystem:         pointerTo("Auto"),
					},
					Suppressed:        true,
					DetachesFromFront: true,
					LateReason: &DisruptionReason{
						TIPLOC:   pointerTo("IJKL"),
						Near:     true,
						ReasonID: 101,
					},
					DisruptionRisk: &ForecastDisruptionRisk{
						Effect: "delay",
						Reason: &DisruptionReason{
							TIPLOC:   pointerTo("MNOP"),
							Near:     true,
							ReasonID: 102,
						},
					},
					AffectedBy: pointerTo("123456"),
					Length:     3,
					PlatformData: &ForecastPlatform{
						Suppressed:      true,
						SuppressedByCIS: true,
						Source:          PlatformDataSourceManual,
						Confirmed:       true,
						Platform:        "2",
					},
				},
			},
		},
	},
	{
		name: "default_values",
		xml: `
		<TS>
			<Location>
				<plat/>
			</Location>
		</TS>
		`,
		expected: ForecastTime{
			Locations: []ForecastLocation{
				{
					PlatformData: &ForecastPlatform{
						Source: PlatformDataSourcePlanned,
					},
				},
			},
		},
	},
}

func TestUnmarshalForecast(t *testing.T) {
	testUnmarshal(t, forecastTestCases)
}
