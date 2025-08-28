package unmarshaller

import (
	"testing"
)

var referenceTestCases = map[string]Reference{
	`<PportTimetableRef timetableId="20050121105940">
		<LocationRef tpl="LEEDS" crs="LDS" toc="RT" locname="Leeds" />
	  <TocRef toc="RT" tocname="Network Rail" url="http://www.nationalrail.co.uk/tocs_maps/tocs/NR.aspx" />
		<LateRunningReasons>
			<Reason code="100" reasontext="This train has been delayed by a broken down train" />
		</LateRunningReasons>
		<CancellationReasons>
			<Reason code="100" reasontext="This train has been cancelled because of a broken down train" />
		</CancellationReasons>
		<CISSource code="at15" name="Leeds LICC" />
		<Via at="BSK" dest="YORK" loc1="COVNTRY" loc2="LEEDS" viatext="via Coventry &amp; Leeds" />
	  <LoadingCategories>
	    <category Code="A" Name="Few seats taken" Toc="">
	      <TypicalDescription>Usually only a few seats taken</TypicalDescription>
	      <ExpectedDescription>Only a few seats taken</ExpectedDescription>
	      <Definition>Everyone will be able to find a seat</Definition>
	      <Colour />
	      <Image />
	    </category>
	    <category Code="B" Name="Plenty of seats" Toc="">
	      <TypicalDescription>Usually plenty of seats available</TypicalDescription>
	      <ExpectedDescription>Plenty of seats available</ExpectedDescription>
	      <Definition>Every will be able to find a seat - groups should still be able to sit together</Definition>
	      <Colour />
	      <Image />
	    </category>
	  </LoadingCategories>
	</PportTimetableRef>`: {
		Locations: []Location{
			{
				Location: "LEEDS",
				CRS:      "LDS",
				TOC:      "RT",
				Name:     "Leeds",
			},
		},
		TrainOperatingCompanies: []TrainOperatingCompany{
			{
				ID:   "RT",
				Name: "Network Rail",
				URL:  "http://www.nationalrail.co.uk/tocs_maps/tocs/NR.aspx",
			},
		},
		LateReasons: []ReasonDescription{
			{
				ReasonID:    100,
				Description: "This train has been delayed by a broken down train",
			},
		},
		CancellationReasons: []ReasonDescription{
			{
				ReasonID:    100,
				Description: "This train has been cancelled because of a broken down train",
			},
		},
		CustomerInformationSystemSources: []CISSource{
			{
				CIS:  "at15",
				Name: "Leeds LICC",
			},
		},
		ViaTexts: []ViaCondition{
			{
				DisplayAt:                "BSK",
				RequiredDestination:      "YORK",
				RequiredCallingLocation1: "COVNTRY",
				RequiredCallingLocation2: "LEEDS",
				Text:                     "via Coventry & Leeds",
			},
		},
		LoadingCategories: []LoadingCategoryReference{
			{
				ID:                  "A",
				Name:                "Few seats taken",
				TypicalDescription:  "Usually only a few seats taken",
				ExpectedDescription: "Only a few seats taken",
				Definition:          "Everyone will be able to find a seat",
			},
			{
				ID:                  "B",
				Name:                "Plenty of seats",
				TypicalDescription:  "Usually plenty of seats available",
				ExpectedDescription: "Plenty of seats available",
				Definition:          "Every will be able to find a seat - groups should still be able to sit together",
			},
		},
	},
}

func TestUnmarshalReference(t *testing.T) {
	if err := TestUnmarshal(referenceTestCases); err != nil {
		t.Error(err)
	}
}
