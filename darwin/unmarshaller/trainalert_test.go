package unmarshaller

import (
	"testing"
)

var trainAlertTestCases = []unmarshalTestCase[TrainAlert]{
	{
		name: "all_fields_are_stored",
		xml: `
		<trainAlert>
			<AlertID>1</AlertID>
			<AlertServices>
				<AlertService RID="012345678901234" UID="A00001" SSD="2025-08-23">
					<Location>LDS</Location>
					<Location>MAN</Location>
				</AlertService>
			</AlertServices>
			<SentAlertBySMS>true</SentAlertBySMS>
			<SentAlertByEmail>true</SentAlertByEmail>
			<SentAlertByTwitter>true</SentAlertByTwitter>
			<Source>NRCC</Source>
			<AlertText>&lt;p&gt;Line 0.&lt;/p&gt;&lt;p&gt;Line 1.&lt;/p&gt;&lt;p&gt;Line 2 and an &lt;a href=&quot;http://example.com&quot;&gt;example link&lt;/a&gt;, an emoji 🇬🇧 &amp;amp; some other&amp;nbsp;text.&lt;/p&gt;</AlertText>
			<Audience>Customer</Audience>
			<AlertType>Normal</AlertType>
			<CopiedFromAlertID>2</CopiedFromAlertID>
			<CopiedFromSource>NT</CopiedFromSource>
		</trainAlert>
		`,
		expected: TrainAlert{
			ID:           "1",
			CopiedFromID: pointerTo("2"),
			Services: []TrainAlertService{
				{
					RID:                pointerTo("012345678901234"),
					UID:                pointerTo("A00001"),
					ScheduledStartDate: pointerTo("2025-08-23"),
					Locations:          []string{"LDS", "MAN"},
				},
			},
			SendSMS:          true,
			SendEmail:        true,
			SendTweet:        true,
			Source:           "NRCC",
			CopiedFromSource: pointerTo("NT"),
			Audience:         "Customer",
			Type:             TrainAlertTypeNormal,
			Message:          `<p>Line 0.</p><p>Line 1.</p><p>Line 2 and an <a href="http://example.com">example link</a>, an emoji 🇬🇧 &amp; some other&nbsp;text.</p>`,
		},
	},
}

func TestUnmarshalTrainAlert(t *testing.T) {
	testUnmarshal(t, trainAlertTestCases)
}
