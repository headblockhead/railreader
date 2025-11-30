package interpreter

import (
	"github.com/google/uuid"
	"github.com/headblockhead/railreader/ingesters/darwin/unmarshaller"
	"github.com/jackc/pgx/v5"
)

func (u UnitOfWork) interpretFormationLoading(formationLoading unmarshaller.FormationLoading) error {
	records, err := u.formationLoadingToRecords(formationLoading)
	if err != nil {
		return err
	}
	err = u.insertFormationLoadingRecords(records)
	if err != nil {
		return err
	}
	return nil
}

type formationLoadingRecord struct {
	ID                           uuid.UUID
	MessageID                    string
	ScheduleID                   string
	FormationID                  string
	LocationID                   string
	ScheduleWorkingArrivalTime   *string
	ScheduleWorkingPassingTime   *string
	ScheduleWorkingDepartureTime *string
	SchedulePublicArrivalTime    *string
	SchedulePublicDepartureTime  *string

	Identifier   string
	Source       *string
	SourceSystem *string
	Percentage   int
}

func (u UnitOfWork) formationLoadingToRecords(formationLoading unmarshaller.FormationLoading) ([]formationLoadingRecord, error) {
	var records []formationLoadingRecord
	for _, fl := range formationLoading.Loading {
		var record formationLoadingRecord
		record.ID = uuid.New()
		record.MessageID = *u.messageID
		record.ScheduleID = formationLoading.RID
		record.FormationID = formationLoading.FormationID
		record.LocationID = formationLoading.TIPLOC
		record.ScheduleWorkingArrivalTime = formationLoading.WorkingArrivalTime
		record.ScheduleWorkingPassingTime = formationLoading.WorkingPassingTime
		record.ScheduleWorkingDepartureTime = formationLoading.WorkingDepartureTime
		record.SchedulePublicArrivalTime = formationLoading.PublicArrivalTime
		record.SchedulePublicDepartureTime = formationLoading.PublicDepartureTime

		record.Identifier = fl.CoachIdentifier
		record.Source = fl.Source
		record.SourceSystem = fl.SourceSystem
		record.Percentage = fl.Percentage

		records = append(records, record)
	}
	return records, nil
}

func (u UnitOfWork) insertFormationLoadingRecords(records []formationLoadingRecord) error {
	batch := &pgx.Batch{}
	for _, record := range records {
		batch.Queue(`
			INSERT INTO darwin.formation_loading (
				id
				,message_id
				,schedule_id
				,formation_id
				,location_id
				,schedule_working_arrival_time
				,schedule_working_passing_time
				,schedule_working_departure_time
				,schedule_public_arrival_time
				,schedule_public_departure_time

				,identifier
				,source
				,source_system
				,percentage
			) VALUES (
				@id
				,@message_id
				,@schedule_id
				,@formation_id
				,@location_id
				,@schedule_working_arrival_time
				,@schedule_working_passing_time
				,@schedule_working_departure_time
				,@schedule_public_arrival_time
				,@schedule_public_departure_time

				,@identifier
				,@source
				,@source_system
				,@percentage
			);
			`, pgx.StrictNamedArgs{
			"id":                              record.ID,
			"message_id":                      record.MessageID,
			"schedule_id":                     record.ScheduleID,
			"formation_id":                    record.FormationID,
			"location_id":                     record.LocationID,
			"schedule_working_arrival_time":   record.ScheduleWorkingArrivalTime,
			"schedule_working_passing_time":   record.ScheduleWorkingPassingTime,
			"schedule_working_departure_time": record.ScheduleWorkingDepartureTime,
			"schedule_public_arrival_time":    record.SchedulePublicArrivalTime,
			"schedule_public_departure_time":  record.SchedulePublicDepartureTime,
			"identifier":                      record.Identifier,
			"source":                          record.Source,
			"source_system":                   record.SourceSystem,
			"percentage":                      record.Percentage,
		})
	}
	results := u.tx.SendBatch(u.ctx, batch)
	return results.Close()
}
