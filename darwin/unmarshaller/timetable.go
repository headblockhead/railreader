package unmarshaller

import "encoding/xml"

// Timetable version 8
type Timetable struct {
	//TODO
	ID string `xml:"timetableID"`
}

func NewTimetable(xmlData string) (tt Timetable, err error) {
	if err = xml.Unmarshal([]byte(xmlData), &tt); err != nil {
		return
	}
	return
}
