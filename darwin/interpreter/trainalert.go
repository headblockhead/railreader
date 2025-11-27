package interpreter

import (
	"github.com/google/uuid"
	"github.com/headblockhead/railreader/darwin/unmarshaller"
	"github.com/jackc/pgx/v5"
)

func (u *UnitOfWork) interpretTrainAlert(trainAlert unmarshaller.TrainAlert) error {
	taRecord, taslRecords, err := u.trainAlertToRecords(trainAlert)
	if err != nil {
		return err
	}
	err = u.insertTrainAlertRecord(taRecord)
	if err != nil {
		return err
	}
	err = u.insertTrainAlertScheduleLocationRecords(taslRecords)
	if err != nil {
		return err
	}
	return nil
}

type trainAlertRecord struct {
	ID uuid.UUID

	MessageID string

	AlertID           string
	CopiedFromAlertID *string
	ShouldSendSMS     bool
	ShouldSendEmail   bool
	ShouldSendTweet   bool
	Source            string
	CopiedFromSource  *string
	Audience          string
	Type              string

	Body string
}

type trainAlertScheduleLocationsRecord struct {
	ID             uuid.UUID
	TrainAlertUUID uuid.UUID
	ScheduleID     string
	LocationIDs    []string
}

func (u *UnitOfWork) trainAlertToRecords(trainAlert unmarshaller.TrainAlert) (trainAlertRecord, []trainAlertScheduleLocationsRecord, error) {
	var taRecord trainAlertRecord
	taRecord.ID = uuid.New()
	taRecord.MessageID = *u.messageID
	taRecord.AlertID = trainAlert.ID
	taRecord.CopiedFromAlertID = trainAlert.CopiedFromID
	taRecord.ShouldSendSMS = trainAlert.SendSMS
	taRecord.ShouldSendEmail = trainAlert.SendEmail
	taRecord.ShouldSendTweet = trainAlert.SendTweet
	taRecord.Source = trainAlert.Source
	taRecord.CopiedFromSource = trainAlert.CopiedFromSource
	taRecord.Audience = string(trainAlert.Audience)
	taRecord.Type = string(trainAlert.Type)
	taRecord.Body = trainAlert.Message

	var taslRecords []trainAlertScheduleLocationsRecord
	for _, tasl := range trainAlert.Services {
		var taslRecord trainAlertScheduleLocationsRecord
		taslRecord.ID = uuid.New()
		taslRecord.TrainAlertUUID = taRecord.ID
		taslRecord.ScheduleID = *tasl.RID // ruh roh
		taslRecord.LocationIDs = tasl.Locations
	}

	return taRecord, taslRecords, nil
}

func (u *UnitOfWork) insertTrainAlertRecord(record trainAlertRecord) error {
	_, err := u.tx.Exec(u.ctx, `
		INSERT INTO darwin.train_alerts (
			id
			,message_id
			,alert_id
			,copied_from_alert_id
			,should_send_sms
			,should_send_email
			,should_send_tweet
			,source
			,copied_from_source
			,audience
			,type
			,body
		) VALUES (
			@id
			,@message_id
			,@alert_id
			,@copied_from_alert_id
			,@should_send_sms
			,@should_send_email
			,@should_send_tweet
			,@source
			,@copied_from_source
			,@audience
			,@type
			,@body
		)
	`, pgx.StrictNamedArgs{
		"id":                   record.ID,
		"message_id":           record.MessageID,
		"alert_id":             record.AlertID,
		"copied_from_alert_id": record.CopiedFromAlertID,
		"should_send_sms":      record.ShouldSendSMS,
		"should_send_email":    record.ShouldSendEmail,
		"should_send_tweet":    record.ShouldSendTweet,
		"source":               record.Source,
		"copied_from_source":   record.CopiedFromSource,
		"audience":             record.Audience,
		"type":                 record.Type,
		"body":                 record.Body,
	})
	if err != nil {
		return err
	}
	return nil
}
