package inserter

import (
	"github.com/google/uuid"
	"github.com/headblockhead/railreader/ingesters/darwin/unmarshaller"
	"github.com/jackc/pgx/v5"
)

func (u *UnitOfWork) insertTrainOrder(trainOrder unmarshaller.TrainOrder) error {
	record, err := u.trainOrderToRecord(trainOrder)
	if err != nil {
		return err
	}
	err = u.insertTrainOrderRecord(record)
	if err != nil {
		return err
	}
	return nil
}

type trainOrderRecord struct {
	ID uuid.UUID

	MessageID string

	LocationID string
	Platform   string

	ClearOrder bool

	FirstServiceRID                  *string
	FirstServiceHeadcode             *string
	FirstServiceWorkingArrivalTime   *string
	FirstServiceWorkingPassingTime   *string
	FirstServiceWorkingDepartureTime *string
	FirstServicePublicArrivalTime    *string
	FirstServicePublicDepartureTime  *string

	SecondServiceRID                  *string
	SecondServiceHeadcode             *string
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

	if trainOrder.Services.First.RIDAndTime != nil {
		record.FirstServiceRID = &trainOrder.Services.First.RIDAndTime.RID
		record.FirstServiceWorkingArrivalTime = trainOrder.Services.First.RIDAndTime.WorkingArrivalTime
		record.FirstServiceWorkingPassingTime = trainOrder.Services.First.RIDAndTime.WorkingPassingTime
		record.FirstServiceWorkingDepartureTime = trainOrder.Services.First.RIDAndTime.WorkingDepartureTime
		record.FirstServicePublicArrivalTime = trainOrder.Services.First.RIDAndTime.PublicArrivalTime
		record.FirstServicePublicDepartureTime = trainOrder.Services.First.RIDAndTime.PublicDepartureTime
	}
	record.FirstServiceHeadcode = trainOrder.Services.First.Headcode

	if trainOrder.Services.Second != nil {
		if trainOrder.Services.Second.RIDAndTime != nil {
			record.SecondServiceRID = &trainOrder.Services.Second.RIDAndTime.RID
			record.SecondServiceWorkingArrivalTime = trainOrder.Services.Second.RIDAndTime.WorkingArrivalTime
			record.SecondServiceWorkingPassingTime = trainOrder.Services.Second.RIDAndTime.WorkingPassingTime
			record.SecondServiceWorkingDepartureTime = trainOrder.Services.Second.RIDAndTime.WorkingDepartureTime
			record.SecondServicePublicArrivalTime = trainOrder.Services.Second.RIDAndTime.PublicArrivalTime
			record.SecondServicePublicDepartureTime = trainOrder.Services.Second.RIDAndTime.PublicDepartureTime
		}
		record.SecondServiceHeadcode = trainOrder.Services.Second.Headcode
	}

	if trainOrder.Services.Third != nil {
		if trainOrder.Services.Third.RIDAndTime != nil {
			record.ThirdServiceRID = &trainOrder.Services.Third.RIDAndTime.RID
			record.ThirdServiceWorkingArrivalTime = trainOrder.Services.Third.RIDAndTime.WorkingArrivalTime
			record.ThirdServiceWorkingPassingTime = trainOrder.Services.Third.RIDAndTime.WorkingPassingTime
			record.ThirdServiceWorkingDepartureTime = trainOrder.Services.Third.RIDAndTime.WorkingDepartureTime
			record.ThirdServicePublicArrivalTime = trainOrder.Services.Third.RIDAndTime.PublicArrivalTime
			record.ThirdServicePublicDepartureTime = trainOrder.Services.Third.RIDAndTime.PublicDepartureTime
		}
		record.ThirdServiceHeadcode = trainOrder.Services.Third.Headcode
	}

	return record, nil
}

func (u *UnitOfWork) insertTrainOrderRecord(record trainOrderRecord) error {
	u.batch.Queue(`
		INSERT INTO darwin.train_orders (
			id
			,message_id
			,location_id
			,platform
			,clear_order
			,first_service_rid
			,first_service_headcode
			,first_service_working_arrival_time
			,first_service_working_passing_time
			,first_service_working_departure_time
			,first_service_public_arrival_time
			,first_service_public_departure_time
			,second_service_rid
			,second_service_headcode
			,second_service_working_arrival_time
			,second_service_working_passing_time
			,second_service_working_departure_time
			,second_service_public_arrival_time
			,second_service_public_departure_time
			,third_service_rid
			,third_service_headcode
			,third_service_working_arrival_time
			,third_service_working_passing_time
			,third_service_working_departure_time
			,third_service_public_arrival_time
			,third_service_public_departure_time
		) VALUES (
			@id
			,@message_id
			,@location_id
			,@platform
			,@clear_order
			,@first_service_rid
			,@first_service_headcode
			,@first_service_working_arrival_time
			,@first_service_working_passing_time
			,@first_service_working_departure_time
			,@first_service_public_arrival_time
			,@first_service_public_departure_time
			,@second_service_rid
			,@second_service_headcode
			,@second_service_working_arrival_time
			,@second_service_working_passing_time
			,@second_service_working_departure_time
			,@second_service_public_arrival_time
			,@second_service_public_departure_time
			,@third_service_rid
			,@third_service_headcode
			,@third_service_working_arrival_time
			,@third_service_working_passing_time
			,@third_service_working_departure_time
			,@third_service_public_arrival_time
			,@third_service_public_departure_time
		)
	`,
		pgx.StrictNamedArgs{
			"id":                                    record.ID,
			"message_id":                            record.MessageID,
			"location_id":                           record.LocationID,
			"platform":                              record.Platform,
			"clear_order":                           record.ClearOrder,
			"first_service_rid":                     record.FirstServiceRID,
			"first_service_headcode":                record.FirstServiceHeadcode,
			"first_service_working_arrival_time":    record.FirstServiceWorkingArrivalTime,
			"first_service_working_passing_time":    record.FirstServiceWorkingPassingTime,
			"first_service_working_departure_time":  record.FirstServiceWorkingDepartureTime,
			"first_service_public_arrival_time":     record.FirstServicePublicArrivalTime,
			"first_service_public_departure_time":   record.FirstServicePublicDepartureTime,
			"second_service_rid":                    record.SecondServiceRID,
			"second_service_headcode":               record.SecondServiceHeadcode,
			"second_service_working_arrival_time":   record.SecondServiceWorkingArrivalTime,
			"second_service_working_passing_time":   record.SecondServiceWorkingPassingTime,
			"second_service_working_departure_time": record.SecondServiceWorkingDepartureTime,
			"second_service_public_arrival_time":    record.SecondServicePublicArrivalTime,
			"second_service_public_departure_time":  record.SecondServicePublicDepartureTime,
			"third_service_rid":                     record.ThirdServiceRID,
			"third_service_headcode":                record.ThirdServiceHeadcode,
			"third_service_working_arrival_time":    record.ThirdServiceWorkingArrivalTime,
			"third_service_working_passing_time":    record.ThirdServiceWorkingPassingTime,
			"third_service_working_departure_time":  record.ThirdServiceWorkingDepartureTime,
			"third_service_public_arrival_time":     record.ThirdServicePublicArrivalTime,
			"third_service_public_departure_time":   record.ThirdServicePublicDepartureTime,
		})
	return nil
}
