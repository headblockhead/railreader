package interpreter

import (
	"github.com/google/uuid"
	"github.com/headblockhead/railreader/ingesters/darwin/unmarshaller"
)

func (u *UnitOfWork) interpretTrainOrder(trainOrder unmarshaller.TrainOrder) error {
	return nil
}

type trainOrderRecord struct {
	ID uuid.UUID

	MessageID string

	LocationID string
	Platform   string

	ClearOrder bool

	FirstServiceRID                  string
	FirstServiceHeadcode             string
	FirstServiceWorkingArrivalTime   *string
	FirstServiceWorkingPassingTime   *string
	FirstServiceWorkingDepartureTime *string
	FirstServicePublicArrivalTime    *string
	FirstServicePublicDepartureTime  *string

	SecondServiceRID                  string
	SecondServiceHeadcode             string
	SecondServiceWorkingArrivalTime   *string
	SecondServiceWorkingPassingTime   *string
	SecondServiceWorkingDepartureTime *string
	SecondServicePublicArrivalTime    *string
	SecondServicePublicDepartureTime  *string

	ThirdServiceRID                  *string
	ThirdServiceHeadcode             *string
	ThirdServiceWorkingArrivalTime   *string
	ThirdServiceWorkingPassingTime   *string
	ThirdServiceWorkingDepartureTime *string
	ThirdServicePublicArrivalTime    *string
	ThirdServicePublicDepartureTime  *string
}

func (u *UnitOfWork) trainOrderToRecord(trainOrder unmarshaller.TrainOrder) (trainOrderRecord, error) {
	var record trainOrderRecord
	record.ID = uuid.New()
	record.MessageID = *u.messageID
	record.LocationID = trainOrder.TIPLOC
	record.Platform = trainOrder.Platform
	record.ClearOrder = bool(trainOrder.ClearOrder)
	if trainOrder.Services == nil {
		return record, nil
	}
	// TODO
}
