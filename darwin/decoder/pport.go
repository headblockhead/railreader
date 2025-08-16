package decoder

import (
	"bytes"
	"encoding/xml"
	"fmt"
)

// PushPort version 18.0
type PushPortMessage struct {
	// Timestamp is in the ISO 8601 YYYY-MM-DDTHH:MM:SS.sssssssss±HH:MM format.
	Timestamp string `xml:"ts,attr"`
	Version   string `xml:"version,attr"`

	// only one of:

	NewTimetableFiles *TimetableFiles `xml:"TimeTableId"`
	StatusUpdate      *Status         `xml:"FailureResp"`
	UpdateResponse    *Response       `xml:"uR"`
	SnapshotResponse  *Response       `xml:"sR"`
}

func NewPushPortMessage(xmlString string) (*PushPortMessage, error) {
	d := xml.NewDecoder(bytes.NewReader([]byte(xmlString)))
	d.Entity = xml.HTMLEntity
	var pport PushPortMessage
	if err := d.Decode(&pport); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message XML: %w", err)
	}
	return &pport, nil
}

// TimetableFiles is sent when there is an update to the timetable reference data, and includes the filenames of the latest versions to be used.
type TimetableFiles struct {
	TimetableFile          string `xml:"ttfile,attr"`
	TimetableReferenceFile string `xml:"ttreffile,attr"`
	// TimeTableId is the exact time the timetable data was written - in the format YYYYMMDDHHMMSS.
	// This is present in the prefix for the filenames of the timetable data, and is provied for convenience.
	TimeTableId string `xml:",chardata"`
}

// Status is a response sent periodically to indicate the status of the system, or to repsond to a bad request.
type Status struct {
	// RequestSourceSystem is optionally provided by the requestor to indicate who they are.
	// This is usually a CISCode.
	RequestSourceSystem string `xml:"requestSource,attr"`
	// RequestID is optionally provided by the requestor to identify their request.
	RequestID string     `xml:"requestID,attr"`
	Code      StatusCode `xml:"code,attr"`

	Description string `xml:",chardata"`
}

type StatusCode string

const (
	StatusCodeOK           StatusCode = "HBOK"
	StatusCodeInitialising StatusCode = "HBINIT"
	StatusCodeFail         StatusCode = "HBFAIL"
	StatusCodeFailover     StatusCode = "HBPENDING"
)

var StatusCodeStrings = map[StatusCode]string{
	StatusCodeOK:           "OK",
	StatusCodeInitialising: "Initialising",
	StatusCodeFail:         "Failure",
	StatusCodeFailover:     "Failover",
}

func (s StatusCode) String() string {
	if str, ok := StatusCodeStrings[s]; ok {
		return str
	}
	return string(s)
}

// Response is a response for a successful request made to update Darwin's data.
// Darwin broadcasts the new state(s) of the data to all subscribers using this message.
type Response struct {
	// Source is optionally provided by the requestor to indicate which system the update originated from (eg "Darwin" or "CIS").
	Source string `xml:"updateOrigin,attr"`
	// SourceSystem is optionally provided by the requestor to indicate who they are. If Source is "CIS", it is most likely a CISCode.
	SourceSystem string `xml:"requestSource,attr"`
	// RequestID is optionally provided by the requestor to identify their request.
	RequestID string `xml:"requestID,attr"`

	// 0 or more of any of these updated elements can be present in a response.
	// This includes 0 of all, which is a valid response.

	Schedules         []Schedule            `xml:"schedule"`
	Deactivations     []Deactivation        `xml:"deactivated"`
	Associations      []Association         `xml:"association"`
	Formations        []FormationsOfService `xml:"scheduleFormations"`
	ForecastTimes     []ForecastTime        `xml:"TS"`
	ServiceLoadings   []ServiceLoading      `xml:"serviceLoading"`
	FormationLoadings []FormationLoading    `xml:"formationLoading"`
	StationMessages   []StationMessage      `xml:"OW"`
	TrainAlerts       []TrainAlert          `xml:"trainAlert"`
	TrainOrders       []TrainOrder          `xml:"trainOrder"`
	HeadcodeChanges   []HeadcodeChange      `xml:"trackingID"`
	Alarms            []Alarm               `xml:"alarm"`
}
