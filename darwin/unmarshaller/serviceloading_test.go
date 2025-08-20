package unmarshaller

import "testing"

var serviceLoadingTestCases = map[string]ServiceLoading{
	`<serviceLoading rid="012345678901234" tpl="ABCD" wta="00:01" wtd="00:02" wtp="00:03" pta="00:04" ptd="00:05">
		<!--I have not seen any examples of the type value yet, the use of "Realtime" here is a complete guess-->
		<loadingCategory type="Realtime" src="CIS" srcInst="CIS1">A</loadingCategory>
		<loadingPercentage type="Realtime" src="CIS" srcInst="CIS1">42</loadingPercentage>
	</serviceLoading>`: {
		RID:    "012345678901234",
		TIPLOC: "ABCD",
		LocationTimeIdentifiers: LocationTimeIdentifiers{
			WorkingArrivalTime:   "00:01",
			WorkingDepartureTime: "00:02",
			WorkingPassingTime:   "00:03",
			PublicArrivalTime:    "00:04",
			PublicDepartureTime:  "00:05",
		},
		LoadingCategory: &LoadingCategory{
			Type:         "Realtime",
			Source:       "CIS",
			SourceSystem: "CIS1",
			Category:     "A",
		},
		LoadingPercentage: &LoadingPercentage{
			Type:         "Realtime",
			Source:       "CIS",
			SourceSystem: "CIS1",
			Percentage:   42,
		},
	},
	// Test default values
	`<serviceLoading>
		<loadingCategory>B</loadingCategory>
		<loadingPercentage>41</loadingPercentage>
	</serviceLoading>`: {
		LoadingCategory: &LoadingCategory{
			Type:     "Typical",
			Category: "B",
		},
		LoadingPercentage: &LoadingPercentage{
			Type:       "Typical",
			Percentage: 41,
		},
	},
}

func TestUnmarshalServiceLoading(t *testing.T) {
	if err := TestUnmarshal(serviceLoadingTestCases); err != nil {
		t.Error(err)
	}
}
