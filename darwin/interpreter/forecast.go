package interpreter

import (
	"github.com/google/uuid"
	"github.com/headblockhead/railreader/darwin/unmarshaller"
	"github.com/jackc/pgx/v5"
)

func (u UnitOfWork) interpretForecast(forecastT unmarshaller.ForecastTime) error {
	// Update schedules record
	sfrecord := u.forecastTimeToRecord(forecastT)
	if err := u.insertScheduleForecastRecord(sfrecord); err != nil {
		return err
	}
	// Update schedule_locations records
	var records []scheduleLocationForecastRecord
	for _, forecastL := range forecastT.Locations {
		record, err := u.forecastTimeLocationToRecord(forecastT.RID, forecastL)
		if err != nil {
			return err
		}
		records = append(records, record)
	}
	return u.insertScheduleLocationForecastRecords(records)
}

type scheduleForecastRecord struct {
	ID uuid.UUID

	MessageID string

	ScheduleID string

	IsReverseFormation bool

	LateReasonID             *int
	LateReasonTIPLOC         *string
	LateReasonIsNearLocation *bool
}

func (u UnitOfWork) forecastTimeToRecord(forecastT unmarshaller.ForecastTime) scheduleForecastRecord {
	var scheduleUpdates scheduleForecastRecord
	scheduleUpdates.ID = uuid.New()
	scheduleUpdates.MessageID = *u.messageID
	scheduleUpdates.ScheduleID = forecastT.RID
	scheduleUpdates.IsReverseFormation = forecastT.ReverseFormation
	if forecastT.LateReason != nil {
		scheduleUpdates.LateReasonID = &forecastT.LateReason.ReasonID
		scheduleUpdates.LateReasonTIPLOC = forecastT.LateReason.TIPLOC
		scheduleUpdates.LateReasonIsNearLocation = &forecastT.LateReason.Near
	}
	return scheduleUpdates
}
func (u UnitOfWork) insertScheduleForecastRecord(record scheduleForecastRecord) error {
	_, err := u.tx.Exec(u.ctx, `
		INSERT INTO darwin.schedule_forecasts (
			id
			,message_id
			,schedule_id
			,is_reverse_formation
			,late_reason_id
			,late_reason_location_id
			,late_reason_is_near_location
		) VALUES (
			@id
			,@message_id
			,@schedule_id
			,@is_reverse_formation
			,@late_reason_id
			,@late_reason_location_id
			,@late_reason_is_near_location
		);
		`, pgx.StrictNamedArgs{
		"id":                           record.ID,
		"message_id":                   record.MessageID,
		"schedule_id":                  record.ScheduleID,
		"is_reverse_formation":         record.IsReverseFormation,
		"late_reason_id":               record.LateReasonID,
		"late_reason_location_id":      record.LateReasonTIPLOC,
		"late_reason_is_near_location": record.LateReasonIsNearLocation,
	})
	return err
}

type scheduleLocationForecastRecord struct {
	ID uuid.UUID

	MessageID string

	ScheduleID                   string
	LocationID                   string
	ScheduleWorkingArrivalTime   *string
	ScheduleWorkingPassingTime   *string
	ScheduleWorkingDepartureTime *string
	SchedulePublicArrivalTime    *string
	SchedulePublicDepartureTime  *string

	ArrivalEstimatedTime          *string
	ArrivalEstimatedWorkingTime   *string
	ArrivalMinimumEstimatedTime   *string
	ArrivalEstimatedTimeIsUnknown *bool
	ArrivalIsDelayed              *bool
	ArrivalActualTime             *string
	ArrivalActualTimeClass        *string
	ArrivalDataSource             *string
	ArrivalDataSourceSystem       *string

	PassingEstimatedTime          *string
	PassingEstimatedWorkingTime   *string
	PassingMinimumEstimatedTime   *string
	PassingEstimatedTimeIsUnknown *bool
	PassingIsDelayed              *bool
	PassingActualTime             *string
	PassingActualTimeClass        *string
	PassingDataSource             *string
	PassingDataSourceSystem       *string

	DepartureEstimatedTime          *string
	DepartureEstimatedWorkingTime   *string
	DepartureMinimumEstimatedTime   *string
	DepartureEstimatedTimeIsUnknown *bool
	DepartureIsDelayed              *bool
	DepartureActualTime             *string
	DepartureActualTimeClass        *string
	DepartureDataSource             *string
	DepartureDataSourceSystem       *string

	LateReasonID             *int
	LateReasonTIPLOC         *string
	LateReasonIsNearLocation *bool

	DisruptionRisk                     *string
	DisruptionRiskReasonID             *int
	DisruptionRiskReasonTIPLOC         *string
	DisruptionRiskReasonIsNearLocation *bool

	AffectedBy                *string
	Length                    *int
	PlatformIsSuppressed      *bool
	PlatformIsSuppressedByCIS *bool
	PlatformDataSource        *string
	PlatformIsConfirmed       *bool
	Platform                  *string
	IsSuppressed              *bool
	DetachesFromFront         *bool
}

func (u UnitOfWork) forecastTimeLocationToRecord(scheduleID string, forecastL unmarshaller.ForecastLocation) (scheduleLocationForecastRecord, error) {
	var record scheduleLocationForecastRecord
	record.ID = uuid.New()
	record.MessageID = *u.messageID
	record.ScheduleID = scheduleID
	record.LocationID = forecastL.TIPLOC
	record.ScheduleWorkingArrivalTime = forecastL.LocationTimeIdentifiers.WorkingArrivalTime
	record.ScheduleWorkingPassingTime = forecastL.LocationTimeIdentifiers.WorkingPassingTime
	record.ScheduleWorkingDepartureTime = forecastL.LocationTimeIdentifiers.WorkingDepartureTime
	record.SchedulePublicArrivalTime = forecastL.LocationTimeIdentifiers.PublicArrivalTime
	record.SchedulePublicDepartureTime = forecastL.LocationTimeIdentifiers.PublicDepartureTime

	if forecastL.ArrivalData != nil {
		record.ArrivalEstimatedTime = forecastL.ArrivalData.EstimatedTime
		record.ArrivalEstimatedWorkingTime = forecastL.ArrivalData.EstimatedWorkingTime
		record.ArrivalMinimumEstimatedTime = forecastL.ArrivalData.EstimatedTimeMinimum
		record.ArrivalEstimatedTimeIsUnknown = &forecastL.ArrivalData.EstimatedTimeUnknown
		record.ArrivalIsDelayed = &forecastL.ArrivalData.Delayed
		record.ArrivalActualTime = forecastL.ArrivalData.ActualTime
		record.ArrivalActualTimeClass = forecastL.ArrivalData.ActualTimeClass
		record.ArrivalDataSource = forecastL.ArrivalData.Source
		record.ArrivalDataSourceSystem = forecastL.ArrivalData.SourceSystem
	}
	if forecastL.PassingData != nil {
		record.PassingEstimatedTime = forecastL.PassingData.EstimatedTime
		record.PassingEstimatedWorkingTime = forecastL.PassingData.EstimatedWorkingTime
		record.PassingMinimumEstimatedTime = forecastL.PassingData.EstimatedTimeMinimum
		record.PassingEstimatedTimeIsUnknown = &forecastL.PassingData.EstimatedTimeUnknown
		record.PassingIsDelayed = &forecastL.PassingData.Delayed
		record.PassingActualTime = forecastL.PassingData.ActualTime
		record.PassingActualTimeClass = forecastL.PassingData.ActualTimeClass
		record.PassingDataSource = forecastL.PassingData.Source
		record.PassingDataSourceSystem = forecastL.PassingData.SourceSystem
	}
	if forecastL.DepartureData != nil {
		record.DepartureEstimatedTime = forecastL.DepartureData.EstimatedTime
		record.DepartureEstimatedWorkingTime = forecastL.DepartureData.EstimatedWorkingTime
		record.DepartureMinimumEstimatedTime = forecastL.DepartureData.EstimatedTimeMinimum
		record.DepartureEstimatedTimeIsUnknown = &forecastL.DepartureData.EstimatedTimeUnknown
		record.DepartureIsDelayed = &forecastL.DepartureData.Delayed
		record.DepartureActualTime = forecastL.DepartureData.ActualTime
		record.DepartureActualTimeClass = forecastL.DepartureData.ActualTimeClass
		record.DepartureDataSource = forecastL.DepartureData.Source
		record.DepartureDataSourceSystem = forecastL.DepartureData.SourceSystem
	}
	if forecastL.LateReason != nil {
		record.LateReasonID = &forecastL.LateReason.ReasonID
		record.LateReasonTIPLOC = forecastL.LateReason.TIPLOC
		record.LateReasonIsNearLocation = &forecastL.LateReason.Near
	}
	if forecastL.DisruptionRisk != nil {
		record.DisruptionRisk = &forecastL.DisruptionRisk.Effect
		if forecastL.DisruptionRisk.Reason != nil {
			record.DisruptionRiskReasonID = &forecastL.DisruptionRisk.Reason.ReasonID
			record.DisruptionRiskReasonTIPLOC = forecastL.DisruptionRisk.Reason.TIPLOC
			record.DisruptionRiskReasonIsNearLocation = &forecastL.DisruptionRisk.Reason.Near
		}
	}
	record.AffectedBy = forecastL.AffectedBy
	record.Length = &forecastL.Length
	if forecastL.PlatformData != nil {
		record.PlatformIsSuppressed = &forecastL.PlatformData.Suppressed
		record.PlatformIsSuppressedByCIS = &forecastL.PlatformData.SuppressedByCIS
		record.PlatformDataSource = &forecastL.PlatformData.Source
		record.PlatformIsConfirmed = &forecastL.PlatformData.Confirmed
		record.Platform = &forecastL.PlatformData.Platform
	}
	record.IsSuppressed = &forecastL.Suppressed
	record.DetachesFromFront = &forecastL.DetachesFromFront
	return record, nil
}

func (u UnitOfWork) insertScheduleLocationForecastRecords(records []scheduleLocationForecastRecord) error {
	batch := &pgx.Batch{}
	for _, record := range records {
		batch.Queue(`
			INSERT INTO darwin.schedule_location_forecasts (
				id
				,message_id
				,schedule_id
				,location_id
				,schedule_working_arrival_time
				,schedule_working_passing_time
				,schedule_working_departure_time
				,schedule_public_arrival_time
				,schedule_public_departure_time
				,arrival_estimated_time
				,arrival_estimated_working_time
				,arrival_minimum_estimated_time
				,arrival_estimated_time_is_unknown
				,arrival_is_delayed
				,arrival_actual_time
				,arrival_actual_time_class
				,arrival_data_source
				,arrival_data_source_system
				,passing_estimated_time
				,passing_estimated_working_time
				,passing_minimum_estimated_time
				,passing_estimated_time_is_unknown
				,passing_is_delayed
				,passing_actual_time
				,passing_actual_time_class
				,passing_data_source
				,passing_data_source_system
				,departure_estimated_time
				,departure_estimated_working_time
				,departure_minimum_estimated_time
				,departure_estimated_time_is_unknown
				,departure_is_delayed
				,departure_actual_time
				,departure_actual_time_class
				,departure_data_source
				,departure_data_source_system
				,late_reason_id
				,late_reason_location_id
				,late_reason_is_near_location
				,disruption_risk
				,disruption_risk_reason_id
				,disruption_risk_reason_location_id
				,disruption_risk_reason_is_near_location
				,affected_by
				,length
				,platform_is_suppressed
				,platform_is_suppressed_by_cis
				,platform_data_source
				,platform_is_confirmed
				,platform
				,is_suppressed
				,detaches_from_front
			) VALUES (
				@id
				,@message_id
				,@schedule_id
				,@location_id
				,@schedule_working_arrival_time
				,@schedule_working_passing_time
				,@schedule_working_departure_time
				,@schedule_public_arrival_time
				,@schedule_public_departure_time
				,@arrival_estimated_time
				,@arrival_estimated_working_time
				,@arrival_minimum_estimated_time
				,@arrival_estimated_time_is_unknown
				,@arrival_is_delayed
				,@arrival_actual_time
				,@arrival_actual_time_class
				,@arrival_data_source
				,@arrival_data_source_system
				,@passing_estimated_time
				,@passing_estimated_working_time
				,@passing_minimum_estimated_time
				,@passing_estimated_time_is_unknown
				,@passing_is_delayed
				,@passing_actual_time
				,@passing_actual_time_class
				,@passing_data_source
				,@passing_data_source_system
				,@departure_estimated_time
				,@departure_estimated_working_time
				,@departure_minimum_estimated_time
				,@departure_estimated_time_is_unknown
				,@departure_is_delayed
				,@departure_actual_time
				,@departure_actual_time_class
				,@departure_data_source
				,@departure_data_source_system
				,@late_reason_id
				,@late_reason_location_id
				,@late_reason_is_near_location
				,@disruption_risk
				,@disruption_risk_reason_id
				,@disruption_risk_reason_location_id
				,@disruption_risk_reason_is_near_location
				,@affected_by
				,@length
				,@platform_is_suppressed
				,@platform_is_suppressed_by_cis
				,@platform_data_source
				,@platform_is_confirmed
				,@platform
				,@is_suppressed
				,@detaches_from_front
			);
			`, pgx.StrictNamedArgs{
			"id":                                      record.ID,
			"message_id":                              record.MessageID,
			"schedule_id":                             record.ScheduleID,
			"location_id":                             record.LocationID,
			"schedule_working_arrival_time":           record.ScheduleWorkingArrivalTime,
			"schedule_working_passing_time":           record.ScheduleWorkingPassingTime,
			"schedule_working_departure_time":         record.ScheduleWorkingDepartureTime,
			"schedule_public_arrival_time":            record.SchedulePublicArrivalTime,
			"schedule_public_departure_time":          record.SchedulePublicDepartureTime,
			"arrival_estimated_time":                  record.ArrivalEstimatedTime,
			"arrival_estimated_working_time":          record.ArrivalEstimatedWorkingTime,
			"arrival_minimum_estimated_time":          record.ArrivalMinimumEstimatedTime,
			"arrival_estimated_time_is_unknown":       record.ArrivalEstimatedTimeIsUnknown,
			"arrival_is_delayed":                      record.ArrivalIsDelayed,
			"arrival_actual_time":                     record.ArrivalActualTime,
			"arrival_actual_time_class":               record.ArrivalActualTimeClass,
			"arrival_data_source":                     record.ArrivalDataSource,
			"arrival_data_source_system":              record.ArrivalDataSourceSystem,
			"passing_estimated_time":                  record.PassingEstimatedTime,
			"passing_estimated_working_time":          record.PassingEstimatedWorkingTime,
			"passing_minimum_estimated_time":          record.PassingMinimumEstimatedTime,
			"passing_estimated_time_is_unknown":       record.PassingEstimatedTimeIsUnknown,
			"passing_is_delayed":                      record.PassingIsDelayed,
			"passing_actual_time":                     record.PassingActualTime,
			"passing_actual_time_class":               record.PassingActualTimeClass,
			"passing_data_source":                     record.PassingDataSource,
			"passing_data_source_system":              record.PassingDataSourceSystem,
			"departure_estimated_time":                record.DepartureEstimatedTime,
			"departure_estimated_working_time":        record.DepartureEstimatedWorkingTime,
			"departure_minimum_estimated_time":        record.DepartureMinimumEstimatedTime,
			"departure_estimated_time_is_unknown":     record.DepartureEstimatedTimeIsUnknown,
			"departure_is_delayed":                    record.DepartureIsDelayed,
			"departure_actual_time":                   record.DepartureActualTime,
			"departure_actual_time_class":             record.DepartureActualTimeClass,
			"departure_data_source":                   record.DepartureDataSource,
			"departure_data_source_system":            record.DepartureDataSourceSystem,
			"late_reason_id":                          record.LateReasonID,
			"late_reason_location_id":                 record.LateReasonTIPLOC,
			"late_reason_is_near_location":            record.LateReasonIsNearLocation,
			"disruption_risk":                         record.DisruptionRisk,
			"disruption_risk_reason_id":               record.DisruptionRiskReasonID,
			"disruption_risk_reason_location_id":      record.DisruptionRiskReasonTIPLOC,
			"disruption_risk_reason_is_near_location": record.DisruptionRiskReasonIsNearLocation,
			"affected_by":                             record.AffectedBy,
			"length":                                  record.Length,
			"platform_is_suppressed":                  record.PlatformIsSuppressed,
			"platform_is_suppressed_by_cis":           record.PlatformIsSuppressedByCIS,
			"platform_data_source":                    record.PlatformDataSource,
			"platform_is_confirmed":                   record.PlatformIsConfirmed,
			"platform":                                record.Platform,
			"is_suppressed":                           record.IsSuppressed,
			"detaches_from_front":                     record.DetachesFromFront,
		})
	}
	results := u.tx.SendBatch(u.ctx, batch)
	return results.Close()
}
