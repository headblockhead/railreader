package interpreter

import (
	"github.com/google/uuid"
	"github.com/headblockhead/railreader/darwin/unmarshaller"
	"github.com/jackc/pgx/v5"
)

func (u UnitOfWork) interpretStationMessage(stationMessage unmarshaller.StationMessage) error {
	record, err := u.stationMessageToRecord(stationMessage)
	if err != nil {
		return err
	}
	err = u.insertStationMessageRecord(record)
	if err != nil {
		return err
	}
	return nil
}

type stationMessageRecord struct {
	ID uuid.UUID

	MessageID string

	StationMessageID string
	Category         string
	Severity         int
	IsSuppressed     bool

	StationCRSCodes []string
	Body            string
}

func (u UnitOfWork) stationMessageToRecord(stationMessage unmarshaller.StationMessage) (stationMessageRecord, error) {
	var record stationMessageRecord
	record.ID = uuid.New()
	record.MessageID = *u.messageID
	record.StationMessageID = stationMessage.ID
	record.Category = string(stationMessage.Category)
	record.Severity = int(stationMessage.Severity)
	record.IsSuppressed = stationMessage.Supressed
	record.StationCRSCodes = make([]string, len(stationMessage.Stations))
	for i, station := range stationMessage.Stations {
		record.StationCRSCodes[i] = station.CRS
	}
	record.Body = stationMessage.Message.Content
	return record, nil
}

func (u UnitOfWork) insertStationMessageRecord(record stationMessageRecord) error {
	_, err := u.tx.Exec(u.ctx, `
		INSERT INTO darwin_station_messages (
			id,
			message_id,
			station_message_id,
			category,
			severity,
			is_suppressed,
			station_crs_codes,
			body
		) VALUES (
			@id,
			@message_id,
			@station_message_id,
			@category,
			@severity,
			@is_suppressed,
			@station_crs_codes,
			@body
		);
	`, pgx.StrictNamedArgs{
		"id":                 record.ID,
		"message_id":         record.MessageID,
		"station_message_id": record.StationMessageID,
		"category":           record.Category,
		"severity":           record.Severity,
		"is_suppressed":      record.IsSuppressed,
		"station_crs_codes":  record.StationCRSCodes,
		"body":               record.Body,
	})
	return err
}
