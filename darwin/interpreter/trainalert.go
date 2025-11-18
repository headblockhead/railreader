package interpreter

import (
	"github.com/google/uuid"
	"github.com/headblockhead/railreader/darwin/unmarshaller"
)

func (u UnitOfWork) interpretTrainAlert(trainAlert unmarshaller.TrainAlert) error {
	taRecord, taslRecords, err := u.trainAlertToRecords(trainAlert)
	if err != nil {
		return err
	}
	if err := u.insertTrainAlertRecord(taRecord); err != nil {
		return err
	}
	if err := u.insertTrainAlertScheduleLocationRecords(taslRecords); err != nil {
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

func (u UnitOfWork) trainAlertToRecords(trainAlert unmarshaller.TrainAlert) (trainAlertRecord, []trainAlertScheduleLocationRecord, error) {
}

func (u UnitOfWork) insertTrainAlertRecord(record trainAlertRecord) error {
}

func (u UnitOfWork) insertTrainAlertScheduleLocationRecords(records []trainAlertScheduleLocationRecord) error {
}
