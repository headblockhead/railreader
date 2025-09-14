package reference

import (
	"testing"

	"github.com/headblockhead/railreader"
)

var referenceTestCases = []unmarshaller.unmarshalTestCase[Reference]{
	{
		name: "all_fields_are_stored",
		xml: `
		<PportTimetableRef timetableId="20050121105940">
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
		    	<category Code="A" Name="Few seats taken">
		    		<TypicalDescription>Usually only a few seats taken</TypicalDescription>
		    		<ExpectedDescription>Only a few seats taken</ExpectedDescription>
		    		<Definition>Everyone will be able to find a seat</Definition>
		    		<Colour />
		    		<Image />
		    	</category>
		    	<category Code="B" Name="Plenty of seats">
		    		<TypicalDescription>Usually plenty of seats available</TypicalDescription>
		    		<ExpectedDescription>Plenty of seats available</ExpectedDescription>
		    		<Definition>Everyone will be able to find a seat - groups should still be able to sit together</Definition>
		    		<Colour />
		    		<Image />
		    	</category>
		  	</LoadingCategories>
		</PportTimetableRef>
		`,
		expected: Reference{
			Locations: []Location{
				{
					Location: "LEEDS",
					CRS:      pointerTo("LDS"),
					TOC:      pointerTo("RT"),
					Name:     "Leeds",
				},
			},
			TrainOperatingCompanies: []TrainOperatingCompany{
				{
					ID:   "RT",
					Name: "Network Rail",
					URL:  pointerTo("http://www.nationalrail.co.uk/tocs_maps/tocs/NR.aspx"),
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
					RequiredCallingLocation2: pointerTo(railreader.TimingPointLocationCode("LEEDS")),
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
					Definition:          "Everyone will be able to find a seat - groups should still be able to sit together",
				},
			},
		},
	},
}

func TestUnmarshalReference(t *testing.T) {
	testUnmarshal(t, referenceTestCases)
}
