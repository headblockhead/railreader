package unmarshaller

import (
	"encoding/xml"

	"github.com/headblockhead/railreader"
)

// Timetable version 8
type Timetable struct {
	ID       string    `xml:"timetableID,attr"`
	Journeys []Journey `xml:"Journey"`
}

type Journey struct {
	TrainIdentifiers

	// TODO fill in more

	// Headcode is the 4-character headcode of the train, with the format:
	// [0-9][A-Z][0-9][0-9]
	Headcode string `xml:"trainId,attr"`
	// TOC is the Rail Delivery Group's 2-character code for the train operating company.
	TOC string `xml:"toc,attr"`
	// Category is a 2-character code for the load of the service.
	// If not provided, it defaults to OO.
	Category railreader.ServiceCategory `xml:"trainCat,attr"`
}

func NewTimetable(xmlData string) (tt Timetable, err error) {
	if err = xml.Unmarshal([]byte(xmlData), &tt); err != nil {
		return
	}
	return
}
