package interpreter

import "github.com/headblockhead/railreader/darwin/unmarshaller"

func timesEqual(a unmarshaller.LocationTimeIdentifiers, b unmarshaller.LocationTimeIdentifiers) bool {
	return a.PublicArrivalTime == b.PublicArrivalTime && a.PublicDepartureTime == b.PublicDepartureTime && a.WorkingArrivalTime == b.WorkingArrivalTime && a.WorkingDepartureTime == b.WorkingDepartureTime && a.WorkingPassingTime == b.WorkingPassingTime
}
