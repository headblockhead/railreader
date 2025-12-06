package unmarshaller

import "testing"

var serviceLoadingTestCases = []unmarshalTestCase[ServiceLoading]{
	{
		name: "all_fields_are_stored",
		xml: `
		<serviceLoading rid="012345678901234" tpl="ABCD" wta="00:01" wtd="00:02" wtp="00:03" pta="00:04" ptd="00:05">
			<!--I have not seen any examples of the type value yet, the use of "Realtime" here is a complete guess-->
			<loadingCategory type="Realtime" src="CIS" srcInst="CIS1">A</loadingCategory>
			<loadingPercentage type="Realtime" src="CIS" srcInst="CIS1">42</loadingPercentage>
		</serviceLoading>
		`,
		expected: ServiceLoading{
			RID:    "012345678901234",
			TIPLOC: "ABCD",
			LocationTimeIdentifiers: LocationTimeIdentifiers{
				WorkingArrivalTime:   pointerTo("00:01"),
				WorkingDepartureTime: pointerTo("00:02"),
				WorkingPassingTime:   pointerTo("00:03"),
				PublicArrivalTime:    pointerTo("00:04"),
				PublicDepartureTime:  pointerTo("00:05"),
			},
			LoadingCategory: &ServiceLoadingCategory{
				Type:         "Realtime",
				Source:       pointerTo("CIS"),
				SourceSystem: pointerTo("CIS1"),
				Category:     "A",
			},
			LoadingPercentage: &ServiceLoadingPercentage{
				Type:         "Realtime",
				Source:       pointerTo("CIS"),
				SourceSystem: pointerTo("CIS1"),
				Percentage:   42,
			},
		},
	},
	{
		name: "default_values",
		xml: `
		<serviceLoading>
			<loadingCategory>B</loadingCategory>
			<loadingPercentage>41</loadingPercentage>
		</serviceLoading>
		`,
		expected: ServiceLoading{
			LoadingCategory: &ServiceLoadingCategory{
				Type:     "Typical",
				Category: "B",
			},
			LoadingPercentage: &ServiceLoadingPercentage{
				Type:       "Typical",
				Percentage: 41,
			},
		},
	},
}

func TestUnmarshalServiceLoading(t *testing.T) {
	testUnmarshal(t, serviceLoadingTestCases)
}
