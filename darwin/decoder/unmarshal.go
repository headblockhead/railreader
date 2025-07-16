package decoder

import (
	"encoding/xml"
	"time"
)

// PushPortMessage is the root node of Darwin messages.
type PushPortMessage struct {
	XMLName xml.Name `xml:"Pport"`

	Timestamp time.Time `xml:"ts,attr"`
	Version   string    `xml:"version,attr"`

	UpdateResponse   *Response `xml:"uR"`
	SnapshotResponse *Response `xml:"sR"`
}

// Response to a request made to update Darwin's data, by broadcasting the new state(s) of the data to all subscribers.
type Response struct {
	// UpdateOrigin is optionally provided by the requestor to indicate which system the update originated from.
	UpdateOrigin string `xml:"updateOrigin,attr"`
	// RequestSource is optionally provided by the requestor to indicate who they are.
	RequestSource string `xml:"requestSource,attr"`
	// RequestID is optionally provided by the requestor to identify their request.
	RequestID string `xml:"requestID,attr"`

	// 0 or more of any of these updated elements can be present in a response.
	// This includes 0 of all, which is a valid response.

	Schedules     []ScheduleInformation     `xml:"schedule"`
	Deactivations []DeactivationInformation `xml:"deactivated"`
	/* Associations                       []AssociationInformation             `xml:"association"`*/
	/*TrainFormations                    []TrainFormationInformation          `xml:"scheduleFormations"`*/
	/*ActualAndForecasts                 []ActualAndForecastInformation       `xml:"TS"`*/
	/*TrainLoadings                      []TrainLoadingInformation            `xml:"formationLoadings"`*/
	/*TableSuppressionAndStationMessages []TableSuppressionAndStationMessages `xml:"OW"`*/
	/*TrainOrders                        []TrainOrderInformation              `xml:"trainOrder"`*/
	/*TrainAlertMessages                 []TrainAlertMessages                 `xml:"trainAlert"`*/
	/*TrackingIDChanges                  []TrackingIDChanges                  `xml:"trackingID"`*/
	/*Alarms                             []Alarms                             `xml:"alarm"`*/
}
