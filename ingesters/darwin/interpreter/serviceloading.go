package interpreter

import (
	"github.com/google/uuid"
	"github.com/headblockhead/railreader/ingesters/darwin/unmarshaller"
	"github.com/jackc/pgx/v5"
)

func (u UnitOfWork) interpretServiceLoading(serviceLoading unmarshaller.ServiceLoading) error {
	record, err := u.serviceLoadingToRecord(serviceLoading)
	if err != nil {
		return err
	}
	err = u.insertServiceLoadingRecord(record)
	if err != nil {
		return err
	}
	return nil
}

type serviceLoadingRecord struct {
	ID uuid.UUID

	MessageID string

	ScheduleID                   string
	LocationID                   string
	ScheduleWorkingArrivalTime   *string
	ScheduleWorkingPassingTime   *string
	ScheduleWorkingDepartureTime *string
	SchedulePublicArrivalTime    *string
	SchedulePublicDepartureTime  *string

	LoadingCategoryCode         *string
	LoadingCategorySource       *string
	LoadingCategorySourceSystem *string
	LoadingCategoryType         *string

	LoadingPercentage             *int
	LoadingPercentageSource       *string
	LoadingPercentageSourceSystem *string
	LoadingPercentageType         *string
}

func (u UnitOfWork) serviceLoadingToRecord(serviceLoading unmarshaller.ServiceLoading) (serviceLoadingRecord, error) {
	var record serviceLoadingRecord
	record.ID = uuid.New()
	record.MessageID = *u.messageID
	record.ScheduleID = serviceLoading.RID
	record.LocationID = serviceLoading.TIPLOC
	record.ScheduleWorkingArrivalTime = serviceLoading.LocationTimeIdentifiers.WorkingArrivalTime
	record.ScheduleWorkingPassingTime = serviceLoading.LocationTimeIdentifiers.WorkingPassingTime
	record.ScheduleWorkingDepartureTime = serviceLoading.LocationTimeIdentifiers.WorkingDepartureTime
	record.SchedulePublicArrivalTime = serviceLoading.LocationTimeIdentifiers.PublicArrivalTime
	record.SchedulePublicDepartureTime = serviceLoading.LocationTimeIdentifiers.PublicDepartureTime
	if serviceLoading.LoadingCategory != nil {
		record.LoadingCategoryCode = &serviceLoading.LoadingCategory.Category
		record.LoadingCategorySource = serviceLoading.LoadingCategory.Source
		record.LoadingCategorySourceSystem = serviceLoading.LoadingCategory.SourceSystem
		record.LoadingCategoryType = (*string)(&serviceLoading.LoadingCategory.Type)
	}
	if serviceLoading.LoadingPercentage != nil {
		record.LoadingPercentage = &serviceLoading.LoadingPercentage.Percentage
		record.LoadingPercentageSource = serviceLoading.LoadingPercentage.Source
		record.LoadingPercentageSourceSystem = serviceLoading.LoadingPercentage.SourceSystem
		record.LoadingPercentageType = (*string)(&serviceLoading.LoadingPercentage.Type)
	}
	return record, nil
}

func (u UnitOfWork) insertServiceLoadingRecord(record serviceLoadingRecord) error {
	_, err := u.tx.Exec(u.ctx, `
		INSERT INTO darwin.service_loadings (
			id,
			message_id,
			schedule_id,
			location_id,
			schedule_working_arrival_time,
			schedule_working_passing_time,
			schedule_working_departure_time,
			schedule_public_arrival_time,
			schedule_public_departure_time,
			loading_category_code,
			loading_category_source,
			loading_category_source_system,
			loading_category_type,
			loading_percentage,
			loading_percentage_source,
			loading_percentage_source_system,
			loading_percentage_type
		) VALUES (
			@id,
			@message_id,
			@schedule_id,
			@location_id,
			@schedule_working_arrival_time,
			@schedule_working_passing_time,
			@schedule_working_departure_time,
			@schedule_public_arrival_time,
			@schedule_public_departure_time,
			@loading_category_code,
			@loading_category_source,
			@loading_category_source_system,
			@loading_category_type,
			@loading_percentage,
			@loading_percentage_source,
			@loading_percentage_source_system,
			@loading_percentage_type
		);
	`, pgx.StrictNamedArgs{
		"id":                               record.ID,
		"message_id":                       record.MessageID,
		"schedule_id":                      record.ScheduleID,
		"location_id":                      record.LocationID,
		"schedule_working_arrival_time":    record.ScheduleWorkingArrivalTime,
		"schedule_working_passing_time":    record.ScheduleWorkingPassingTime,
		"schedule_working_departure_time":  record.ScheduleWorkingDepartureTime,
		"schedule_public_arrival_time":     record.SchedulePublicArrivalTime,
		"schedule_public_departure_time":   record.SchedulePublicDepartureTime,
		"loading_category_code":            record.LoadingCategoryCode,
		"loading_category_source":          record.LoadingCategorySource,
		"loading_category_source_system":   record.LoadingCategorySourceSystem,
		"loading_category_type":            record.LoadingCategoryType,
		"loading_percentage":               record.LoadingPercentage,
		"loading_percentage_source":        record.LoadingPercentageSource,
		"loading_percentage_source_system": record.LoadingPercentageSourceSystem,
		"loading_percentage_type":          record.LoadingPercentageType,
	})
	if err != nil {
		return err
	}
	return nil
}
