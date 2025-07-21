package decoder

import (
	"encoding/xml"
)

// PushPortMessage is the root node of Darwin messages.
type PushPortMessage struct {
	XMLName xml.Name `xml:"Pport"`

	Timestamp string `xml:"ts,attr"`
	Version   string `xml:"version,attr"`

	// Message contains only one of:

	TimeTableID     *TimeTableId `xml:"TimeTableId"`
	FailureResponse *Status      `xml:"FailureResp"`
	// SnapshotID is the optionally provided ID of a snapshot file that can be downloaded via FTP.
	SnapshotID       *string   `xml:"SnapshotId"`
	UpdateResponse   *Response `xml:"uR"`
	SnapshotResponse *Response `xml:"sR"`
}

// TimeTableId is sent when there is an update to the timetable reference data, and includes the filenames of the latest versions to be used.
type TimeTableId struct {
	// TTfile is the optionally provided filename of the lastest version of the timetable data.
	// Timetable data provides a list of all trains that are scheduled to run in the timetable, and their associations.
	TTfile *string `xml:"ttfile,attr"`
	// TTRefFile is the optionally provided filename of the latest version of the reference data.
	// Reference data provides a list of data about:
	// all possible TIPLOCs,
	// most Train Operating Compaines,
	// all reasons for trains to run late,
	// all reasons for trains to be cancelled,
	// all possible conditions to add 'via' to a train's signage,
	// most CIS sources (source that provided the train info - the CISCode type),
	// and categories for train loading amounts (possibly unused?).
	TTRefFile *string `xml:"ttreffile,attr"`
	// TimeTableId is the exact time the timetable data was written - in the format YYYYMMDDHHMMSS.
	// This is present in the prefix for the filenames of the timetable data, and is provied for convenience.
	TimeTableId *string `xml:",chardata"`
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

// CISCode is a code that identifies the ID of the system that sent the request.
// A mapping of CIS codes to system names is included in the reference data.
type CISCode string

// Status is a response sent periodically to indicate the status of the system, or to repsond to a bad request.
type Status struct {
	// RequestSourceSystem is optionally provided by the requestor to indicate who they are.
	RequestSourceSystem *CISCode `xml:"requestSource,attr"`
	// RequestID is optionally provided by the requestor to identify their request.
	RequestID *string    `xml:"requestID,attr"`
	Code      StatusCode `xml:"code,attr"`
}

// Response is a response for a successful request made to update Darwin's data.
// Darwin broadcasts the new state(s) of the data to all subscribers using this message.
type Response struct {
	// UpdateOrigin is optionally provided by the requestor to indicate which system the update originated from.
	UpdateOrigin *string `xml:"updateOrigin,attr"`
	// RequestSource is optionally provided by the requestor to indicate who they are.
	RequestSource *string `xml:"requestSource,attr"`
	// RequestID is optionally provided by the requestor to identify their request.
	RequestID *string `xml:"requestID,attr"`

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
	/*TrainOrders                        []TrainOrderInformation              `xml:"trainOrder"`*/
	/*TrainAlertMessages                 []TrainAlertMessages                 `xml:"trainAlert"`*/
	/*TrackingIDChanges                  []TrackingIDChanges                  `xml:"trackingID"`*/
	/*Alarms                             []Alarms                             `xml:"alarm"`*/
}
