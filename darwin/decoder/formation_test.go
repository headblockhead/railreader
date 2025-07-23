package decoder

import "testing"

var formationTestCases = map[string]FormationsOfService{
	`<scheduleFormations rid="012345678901234">
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
	</scheduleFormations>`: {
		RID: "012345678901234",
		Formations: []Formation{
			{
				ID:           "012345678901234-001",
				Source:       "CIS",
				SourceSystem: "CIS1",
				Coaches: []Coach{
					{
						CoachIdentifier: "A",
						CoachClass:      "First",
						Toilet: ToiletInformation{
							ToiletStatus: ToiletStatusNotInService,
							ToiletType:   ToiletTypeStandard,
						},
					},
					{
						CoachIdentifier: "B",
						Toilet: ToiletInformation{
							ToiletStatus: ToiletStatusInService,
							ToiletType:   ToiletTypeUnknown,
						},
					},
				},
			},
			{
				ID: "012345678901234-002",
				Coaches: []Coach{
					{
						CoachIdentifier: "1",
					},
				},
			},
		},
	},
}

func TestUnmarshalFormation(t *testing.T) {
	if err := TestUnmarshal(formationTestCases); err != nil {
		t.Errorf("%v", err)
	}
}
