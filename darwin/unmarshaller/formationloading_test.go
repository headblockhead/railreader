package unmarshaller

import "testing"

var formationLoadingTestCases = []unmarshalTestCase[FormationLoading]{
	{
		name: "all_fields_are_stored",
		xml: `
		<formationLoading fid="012345678901234-001" rid="012345678901234" tpl="ABCD" wta="00:01" wtd="00:02" wtp="00:03" pta="00:04" ptd="00:05">
			<loading coachNumber="A" src="CIS" srcInst="CIS1">40</loading>
			<loading coachNumber="B" src="CIS" srcInst="CIS1">23</loading>
			<loading coachNumber="C"/>
		</formationLoading>
		`,
		expected: FormationLoading{
			FormationID: "012345678901234-001",
			RID:         "012345678901234",
			TIPLOC:      "ABCD",
			LocationTimeIdentifiers: LocationTimeIdentifiers{
				PublicArrivalTime:    pointerTo("00:04"),
				PublicDepartureTime:  pointerTo("00:05"),
				WorkingArrivalTime:   pointerTo("00:01"),
				WorkingDepartureTime: pointerTo("00:02"),
				WorkingPassingTime:   pointerTo("00:03"),
			},
			Loading: []CoachLoading{
				{
					CoachIdentifier: "A",
					Source:          pointerTo("CIS"),
					SourceSystem:    pointerTo("CIS1"),
					Percentage:      40,
				},
				{
					CoachIdentifier: "B",
					Source:          pointerTo("CIS"),
					SourceSystem:    pointerTo("CIS1"),
					Percentage:      23,
				},
				{
					CoachIdentifier: "C",
				},
			},
		},
	},
}

func TestUnmarshalFormationLoading(t *testing.T) {
	testUnmarshal(t, formationLoadingTestCases)
}
