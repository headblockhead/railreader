package decoder

import "testing"

var formationLoadingTestCases = map[string]FormationLoading{
	`<formationLoading fid="012345678901234-001" rid="012345678901234" tpl="ABCD" wta="00:01" wtd="00:02" wtp="00:03" pta="00:04" ptd="00:05">
		<loading coachNumber="A" src="CIS" srcInst="CIS1">40</loading>
		<loading coachNumber="B" src="CIS" srcInst="CIS1">23</loading>
		<loading coachNumber="C"/>
	</formationLoading>`: {
		FormationID:          "012345678901234-001",
		RID:                  "012345678901234",
		TIPLOC:               "ABCD",
		PublicArrivalTime:    "00:04",
		PublicDepartureTime:  "00:05",
		WorkingArrivalTime:   "00:01",
		WorkingDepartureTime: "00:02",
		WorkingPassingTime:   "00:03",
		Loading: []CoachLoadingData{
			{
				CoachIdentifier: "A",
				Source:          "CIS",
				SourceSystem:    "CIS1",
				Percentage:      40,
			},
			{
				CoachIdentifier: "B",
				Source:          "CIS",
				SourceSystem:    "CIS1",
				Percentage:      23,
			},
			{
				CoachIdentifier: "C",
			},
		},
	},
}

func TestUnmarshalFormationLoading(t *testing.T) {
	if err := TestUnmarshal(formationLoadingTestCases); err != nil {
		t.Error(err)
	}
}
