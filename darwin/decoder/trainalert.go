package decoder

import "github.com/headblockhead/railreader"

type TrainAlert struct {
	ID string `xml:"AlertID"`
	// CopiedFromID is the (optionally provided) ID of the alert that this alert was copied from.
	CopiedFromID *string `xml:"CopiedFromAlertID,attr"`
	// Services can be empty.
	Services  []TrainAlertService `xml:"AlertServices>AlertService"`
	SendSMS   bool                `xml:"SentAlertBySMS,attr"`
	SendEmail bool                `xml:"SentAlertByEmail,attr"`
	SendTweet bool                `xml:"SentAlertByTwitter,attr"`
	// Source is usually a TOC code, but can also be "NRCC" (the National Rail Communications Centre?).
	Source string `xml:"Source,attr"`
	// CopiedFromSource is the (optionally provided) source of the alert that this alert was copied from.
	CopiedFromSource *string `xml:"CopiedFromSource,attr"`
	// Audience is usually "Customer", but may be other values.
	Audience string `xml:"Audience,attr"`
	// All other tags are interpreted as part of the message body.
	Message XHTMLBody `xml:",any"`
}

type TrainAlertService struct {
	// zero or more of:

	// RID is the unique 16-character ID for a specific train.
	RID *string `xml:"RID,attr"`
	// UID is (despite the name) a non-unique 6-character ID for a route at a time of day.
	UID *string `xml:"UID,attr"`
	// ScheduledStartDate in YYYY-MM-DD format.
	ScheduledStartDate *string `xml:"SSD,attr"`

	// Locations /should/ contain at least one Location element.
	Locations []railreader.TIPLOC `xml:"Location"`
}
