package decoder

import (
	"testing"

	"github.com/headblockhead/railreader"
)

var trainAlertTestCases = map[string]TrainAlert{
	`<trainAlert>
		<AlertID>1</AlertID>
		<AlertServices>
			<AlertService RID="012345678901234" UID="A00001" SSD="2006-01-02">
				<Location>LDS</Location>
				<Location>MAN</Location>
			</AlertService>
		</AlertServices>
		<SentAlertBySMS>true</SentAlertBySMS>
		<SentAlertByEmail>true</SentAlertByEmail>
		<SentAlertByTwitter>true</SentAlertByTwitter>
		<Source>NRCC</Source>
		<AlertText>&lt;p&gt;Line 0.&lt;/p&gt;&lt;p&gt;Line 1.&lt;/p&gt;&lt;p&gt;Line 2 and an &lt;a href="http://example.com"&gt;example link&lt;/a&gt; and some other text.&lt;/p&gt;</AlertText>
		<Audience>Customer</Audience>
		<AlertType>Normal</AlertType>
		<CopiedFromAlertID>2</CopiedFromAlertID>
		<CopiedFromSource>NT</CopiedFromSource>
	</trainAlert>`: {
		ID:           "1",
		CopiedFromID: "2",
		Services: []TrainAlertService{
			{
				RID:                "012345678901234",
				UID:                "A00001",
				ScheduledStartDate: "2006-01-02",
				Locations:          []railreader.TIPLOC{"LDS", "MAN"},
			},
		},
		SendSMS:          true,
		SendEmail:        true,
		SendTweet:        true,
		Source:           "NRCC",
		CopiedFromSource: "NT",
		Audience:         "Customer",
		Type:             TrainAlertTypeNormal,
		Message:          `<p>Line 0.</p><p>Line 1.</p><p>Line 2 and an <a href="http://example.com">example link</a> and some other text.</p>`,
	},
}

func TestUnmarshalTrainAlert(t *testing.T) {
	if err := TestUnmarshal(trainAlertTestCases); err != nil {
		t.Error(err)
	}
}
