package unmarshaller

import (
	"github.com/headblockhead/railreader"
)

type TrainAlert struct {
	ID string `xml:"AlertID"`
	// CopiedFromID is the (optionally provided) ID value of the alert that this alert was copied from.
	CopiedFromID *string `xml:"CopiedFromAlertID"`
	// Services can be empty, however if this is the case, the message shouldn't be shown.
	Services  []TrainAlertService `xml:"AlertServices>AlertService"`
	SendSMS   bool                `xml:"SentAlertBySMS"`
	SendEmail bool                `xml:"SentAlertByEmail"`
	SendTweet bool                `xml:"SentAlertByTwitter"`
	// Source is usually a TOC code, but can also be "NRCC" (the National Rail Communications Centre?).
	Source string `xml:"Source"`
	// CopiedFromSource is the (optionally provided) Source value of the alert that this alert was copied from.
	CopiedFromSource *string            `xml:"CopiedFromSource"`
	Audience         TrainAlertAudience `xml:"Audience"`
	Type             TrainAlertType     `xml:"AlertType"`
	// Message is a basic HTML string. (containing only <p> and <a> tags).
	Message string `xml:"AlertText"`
}

type TrainAlertService struct {
	// TrainIdentifiers is not used here, because for some reason the RID UID and SSD attributes are uniquely capitalised here.

	// (most likely) at least one of:
	RID                *string `xml:"RID,attr"`
	UID                *string `xml:"UID,attr"`
	ScheduledStartDate *string `xml:"SSD,attr"`
	// to identify a specific service.

	// Locations is the list of TIPLOCs that this alert applies to.
	Locations []railreader.TimingPointLocationCode `xml:"Location"`
}

type TrainAlertType string

const (
	TrainAlertTypeNormal TrainAlertType = "Normal"
	TrainAlertTypeForced TrainAlertType = "Forced"
)

type TrainAlertAudience string

const (
	TrainAlertAudienceCustomer   TrainAlertAudience = "Customer"
	TrainAlertAudienceStaff      TrainAlertAudience = "Staff"
	TrainAlertAudienceOperations TrainAlertAudience = "Operations"
)
