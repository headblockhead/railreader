package inserter

import (
	"github.com/google/uuid"
	"github.com/headblockhead/railreader/ingesters/darwin/unmarshaller"
	"github.com/jackc/pgx/v5"
)

func (u *UnitOfWork) insertForecast(forecastT unmarshaller.ForecastTime) error {
	sfRecord, sflRecords, err := u.forecastTimeToRecords(forecastT)
	if err != nil {
		return err
	}
	err = u.insertForecastRecord(sfRecord)
	if err != nil {
		return err
	}
	err = u.insertForecastLocationRecords(sflRecords)
	if err != nil {
		return err
	}
	return nil
}

type forecastRecord struct {
	ID uuid.UUID

	MessageID string

	ScheduleID string

	IsReverseFormation bool

	LateReasonID             *int
	LateReasonTIPLOC         *string
	LateReasonIsNearLocation *bool
}

type forecastLocationRecord struct {
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

func (u *UnitOfWork) forecastTimeToRecords(forecastT unmarshaller.ForecastTime) (forecastRecord, []forecastLocationRecord, error) {
	var sfRecord forecastRecord
	sfRecord.ID = uuid.New()
	sfRecord.MessageID = *u.messageID
	sfRecord.ScheduleID = forecastT.RID
	sfRecord.IsReverseFormation = forecastT.ReverseFormation
	if forecastT.LateReason != nil {
		sfRecord.LateReasonID = &forecastT.LateReason.ReasonID
		sfRecord.LateReasonTIPLOC = forecastT.LateReason.TIPLOC
		sfRecord.LateReasonIsNearLocation = &forecastT.LateReason.Near
	}

	var sflRecords []forecastLocationRecord
	for _, forecastL := range forecastT.Locations {
		var sflRecord forecastLocationRecord
		sflRecord.ID = uuid.New()
		sflRecord.MessageID = *u.messageID
		sflRecord.ScheduleID = forecastT.RID
		sflRecord.LocationID = forecastL.TIPLOC
		sflRecord.ScheduleWorkingArrivalTime = forecastL.LocationTimeIdentifiers.WorkingArrivalTime
		sflRecord.ScheduleWorkingPassingTime = forecastL.LocationTimeIdentifiers.WorkingPassingTime
		sflRecord.ScheduleWorkingDepartureTime = forecastL.LocationTimeIdentifiers.WorkingDepartureTime
		sflRecord.SchedulePublicArrivalTime = forecastL.LocationTimeIdentifiers.PublicArrivalTime
		sflRecord.SchedulePublicDepartureTime = forecastL.LocationTimeIdentifiers.PublicDepartureTime

		if forecastL.ArrivalData != nil {
			sflRecord.ArrivalEstimatedTime = forecastL.ArrivalData.EstimatedTime
			sflRecord.ArrivalEstimatedWorkingTime = forecastL.ArrivalData.EstimatedWorkingTime
			sflRecord.ArrivalMinimumEstimatedTime = forecastL.ArrivalData.EstimatedTimeMinimum
			sflRecord.ArrivalEstimatedTimeIsUnknown = &forecastL.ArrivalData.EstimatedTimeUnknown
			sflRecord.ArrivalIsDelayed = &forecastL.ArrivalData.Delayed
			sflRecord.ArrivalActualTime = forecastL.ArrivalData.ActualTime
			sflRecord.ArrivalActualTimeClass = forecastL.ArrivalData.ActualTimeClass
			sflRecord.ArrivalDataSource = forecastL.ArrivalData.Source
			sflRecord.ArrivalDataSourceSystem = forecastL.ArrivalData.SourceSystem
		}
		if forecastL.PassingData != nil {
			sflRecord.PassingEstimatedTime = forecastL.PassingData.EstimatedTime
			sflRecord.PassingEstimatedWorkingTime = forecastL.PassingData.EstimatedWorkingTime
			sflRecord.PassingMinimumEstimatedTime = forecastL.PassingData.EstimatedTimeMinimum
			sflRecord.PassingEstimatedTimeIsUnknown = &forecastL.PassingData.EstimatedTimeUnknown
			sflRecord.PassingIsDelayed = &forecastL.PassingData.Delayed
			sflRecord.PassingActualTime = forecastL.PassingData.ActualTime
			sflRecord.PassingActualTimeClass = forecastL.PassingData.ActualTimeClass
			sflRecord.PassingDataSource = forecastL.PassingData.Source
			sflRecord.PassingDataSourceSystem = forecastL.PassingData.SourceSystem
		}
		if forecastL.DepartureData != nil {
			sflRecord.DepartureEstimatedTime = forecastL.DepartureData.EstimatedTime
			sflRecord.DepartureEstimatedWorkingTime = forecastL.DepartureData.EstimatedWorkingTime
			sflRecord.DepartureMinimumEstimatedTime = forecastL.DepartureData.EstimatedTimeMinimum
			sflRecord.DepartureEstimatedTimeIsUnknown = &forecastL.DepartureData.EstimatedTimeUnknown
			sflRecord.DepartureIsDelayed = &forecastL.DepartureData.Delayed
			sflRecord.DepartureActualTime = forecastL.DepartureData.ActualTime
			sflRecord.DepartureActualTimeClass = forecastL.DepartureData.ActualTimeClass
			sflRecord.DepartureDataSource = forecastL.DepartureData.Source
			sflRecord.DepartureDataSourceSystem = forecastL.DepartureData.SourceSystem
		}
		if forecastL.LateReason != nil {
			sflRecord.LateReasonID = &forecastL.LateReason.ReasonID
			sflRecord.LateReasonTIPLOC = forecastL.LateReason.TIPLOC
			sflRecord.LateReasonIsNearLocation = &forecastL.LateReason.Near
		}
		if forecastL.DisruptionRisk != nil {
			sflRecord.DisruptionRisk = &forecastL.DisruptionRisk.Effect
			if forecastL.DisruptionRisk.Reason != nil {
				sflRecord.DisruptionRiskReasonID = &forecastL.DisruptionRisk.Reason.ReasonID
				sflRecord.DisruptionRiskReasonTIPLOC = forecastL.DisruptionRisk.Reason.TIPLOC
				sflRecord.DisruptionRiskReasonIsNearLocation = &forecastL.DisruptionRisk.Reason.Near
			}
		}
		sflRecord.AffectedBy = forecastL.AffectedBy
		sflRecord.Length = &forecastL.Length
		if forecastL.PlatformData != nil {
			sflRecord.PlatformIsSuppressed = &forecastL.PlatformData.Suppressed
			sflRecord.PlatformIsSuppressedByCIS = &forecastL.PlatformData.SuppressedByCIS
			sflRecord.PlatformDataSource = &forecastL.PlatformData.Source
			sflRecord.PlatformIsConfirmed = &forecastL.PlatformData.Confirmed
			sflRecord.Platform = &forecastL.PlatformData.Platform
		}
		sflRecord.IsSuppressed = &forecastL.Suppressed
		sflRecord.DetachesFromFront = &forecastL.DetachesFromFront
		sflRecords = append(sflRecords, sflRecord)
	}
	return sfRecord, sflRecords, nil
}

func (u *UnitOfWork) insertForecastRecord(record forecastRecord) error {
	u.batch.Queue(`
		INSERT INTO darwin.forecasts (
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
	return nil
}

func (u *UnitOfWork) insertForecastLocationRecords(records []forecastLocationRecord) error {
	for _, record := range records {
		u.batch.Queue(`
			INSERT INTO darwin.forecast_locations (
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
	return nil
}
