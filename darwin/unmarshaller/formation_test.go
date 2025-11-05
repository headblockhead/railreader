package unmarshaller

import "testing"

var formationTestCases = []unmarshalTestCase[FormationsOfService]{
	{
		name: "all_fields_are_stored",
		xml: `
		<scheduleFormations rid="012345678901234">
			<formation fid="012345678901234-001" src="CIS" srcInst="CIS1">
				<coaches>
					<coach coachNumber="A" coachClass="First">
						<toilet status="NotInService">Standard</toilet>
					</coach>
					<coach coachNumber="B">
						<!--This should default to an unknown toilet in service-->
						<toilet/>
					</coach>
				</coaches>
			</formation>
			<formation fid="012345678901234-002">
				<coaches>
					<coach coachNumber="1"/>
				</coaches>
			</formation>
		</scheduleFormations>
		`,
		expected: FormationsOfService{
			RID: "012345678901234",
			Formations: []Formation{
				{
					ID:           "012345678901234-001",
					Source:       pointerTo("CIS"),
					SourceSystem: pointerTo("CIS1"),
					Coaches: []FormationCoach{
						{
							Identifier: "A",
							Class:      pointerTo("First"),
							Toilet: FormationCoachToilet{
								Status: "NotInService",
								Type:   "Standard",
							},
						},
						{
							Identifier: "B",
							Toilet: FormationCoachToilet{
								Status: "InService",
								Type:   "Unknown",
							},
						},
					},
				},
				{
					ID: "012345678901234-002",
					Coaches: []FormationCoach{
						{
							Identifier: "1",
						},
					},
				},
			},
		},
	},
}

func TestUnmarshalFormation(t *testing.T) {
	testUnmarshal(t, formationTestCases)
}
