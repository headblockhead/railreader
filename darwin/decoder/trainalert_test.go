package decoder

import "testing"

var trainAlertTestCases = map[string]TrainAlert{
	`<trainAlert>
		<AlertID>1</AlertID>
		<Services>
			<Service UID="uidhere" SSD="ssdhere">
				<Location>LDS</Location>
				<Location>MAN</Location>
			</Service>
		</Services>
	</trainAlert>`: {},
}

func TestUnmarshalTrainAlert(t *testing.T) {
	if err := TestUnmarshal(trainAlertTestCases); err != nil {
		t.Error(err)
	}
}
