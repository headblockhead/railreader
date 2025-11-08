package interpreter

import (
	"github.com/headblockhead/railreader/darwin/unmarshaller"
	"github.com/jackc/pgx/v5"
)

func (u UnitOfWork) interpretForecast(forecastT unmarshaller.ForecastTime) error {
	// Update schedules record
	scheduleUpdates := u.forecastTimeToScheduleUpdates(forecastT)
	if err := u.updateOneScheduleWithForecastUpdates(scheduleUpdates); err != nil {
		return err
	}
	// Update schedule_locations records
	for _, forecastL := range forecastT.Locations {
		sequence, err := u.findLocationSequence(forecastT.RID, forecastL.TIPLOC, forecastL.LocationTimeIdentifiers)
		if err != nil {
			return err
		}
		scheduleLocationUpdates := u.forecastTimeLocationToScheduleLocationUpdates(forecastT.RID, sequence, forecastL)
		if err := u.updateOneScheduleLocationWithForecastUpdates(scheduleLocationUpdates); err != nil {
			return err
		}
	}
	return nil
}

type forecastScheduleUpdates struct {
	ScheduleID string

	IsReverseFormation       bool
	LateReasonID             *int
	LateReasonTIPLOC         *string
	LateReasonIsNearLocation *bool
}

func (u UnitOfWork) forecastTimeToScheduleUpdates(forecastT unmarshaller.ForecastTime) forecastScheduleUpdates {
	var scheduleUpdates forecastScheduleUpdates
	scheduleUpdates.ScheduleID = forecastT.RID
	scheduleUpdates.IsReverseFormation = forecastT.ReverseFormation
	if forecastT.LateReason != nil {
		scheduleUpdates.LateReasonID = &forecastT.LateReason.ReasonID
		scheduleUpdates.LateReasonTIPLOC = forecastT.LateReason.TIPLOC
		scheduleUpdates.LateReasonIsNearLocation = &forecastT.LateReason.Near
	}
	return scheduleUpdates
}
func (u UnitOfWork) updateOneScheduleWithForecastUpdates(updates forecastScheduleUpdates) error {
	_, err := u.tx.Exec(u.ctx, `
		UPDATE darwin.schedules SET 
			is_reverse_formation = @is_reverse_formation
			,late_reason_id = @late_reason_id
			,late_reason_location_id = @late_reason_location_id
			,late_reason_is_near_location = @late_reason_is_near_location
		WHERE schedule_id = @schedule_id;
		`, pgx.StrictNamedArgs{
		"schedule_id":                  updates.ScheduleID,
		"is_reverse_formation":         updates.IsReverseFormation,
		"late_reason_id":               updates.LateReasonID,
		"late_reason_location_id":      updates.LateReasonTIPLOC,
		"late_reason_is_near_location": updates.LateReasonIsNearLocation,
	})
	return err
}

type forecastScheduleLocationUpdates struct {
	ScheduleID string
	Sequence   int

	ArrivalEstimatedTime               *string
	ArrivalEstimatedWorkingTime        *string
	ArrivalMinimumEstimatedTime        *string
	ArrivalEstimatedTimeIsUnknown      *bool
	ArrivalIsDelayed                   *bool
	ArrivalActualTime                  *string
	ArrivalActualTimeClass             *string
	ArrivalDataSource                  *string
	ArrivalDataSourceSystem            *string
	PassingEstimatedTime               *string
	PassingEstimatedWorkingTime        *string
	PassingMinimumEstimatedTime        *string
	PassingEstimatedTimeIsUnknown      *bool
	PassingIsDelayed                   *bool
	PassingActualTime                  *string
	PassingActualTimeClass             *string
	PassingDataSource                  *string
	PassingDataSourceSystem            *string
	DepartureEstimatedTime             *string
	DepartureEstimatedWorkingTime      *string
	DepartureMinimumEstimatedTime      *string
	DepartureEstimatedTimeIsUnknown    *bool
	DepartureIsDelayed                 *bool
	DepartureActualTime                *string
	DepartureActualTimeClass           *string
	DepartureDataSource                *string
	DepartureDataSourceSystem          *string
	LateReasonID                       *int
	LateReasonTIPLOC                   *string
	LateReasonIsNearLocation           *bool
	DisruptionRisk                     *string
	DisruptionRiskReasonID             *int
	DisruptionRiskReasonTIPLOC         *string
	DisruptionRiskReasonIsNearLocation *bool
	AffectedBy                         *string
	Length                             *int
	PlatformIsSuppressed               *bool
	PlatformIsSuppressedByCIS          *bool
	PlatformDataSource                 *string
	PlatformConfirmed                  *bool
	Platform                           *string
	IsSuppressed                       *bool
	DetachesFromFront                  *bool
}

func (u UnitOfWork) forecastTimeLocationToScheduleLocationUpdates(scheduleID string, sequence int, forecastL unmarshaller.ForecastLocation) forecastScheduleLocationUpdates {
	var updates forecastScheduleLocationUpdates
	updates.ScheduleID = scheduleID
	updates.Sequence = sequence
	if forecastL.ArrivalData != nil {
		updates.ArrivalEstimatedTime = forecastL.ArrivalData.EstimatedTime
		updates.ArrivalEstimatedWorkingTime = forecastL.ArrivalData.EstimatedWorkingTime
		updates.ArrivalMinimumEstimatedTime = forecastL.ArrivalData.EstimatedTimeMinimum
		updates.ArrivalEstimatedTimeIsUnknown = &forecastL.ArrivalData.EstimatedTimeUnknown
		updates.ArrivalIsDelayed = &forecastL.ArrivalData.Delayed
		updates.ArrivalActualTime = forecastL.ArrivalData.ActualTime
		updates.ArrivalActualTimeClass = forecastL.ArrivalData.ActualTimeClass
		updates.ArrivalDataSource = forecastL.ArrivalData.Source
		updates.ArrivalDataSourceSystem = forecastL.ArrivalData.SourceSystem
	}
	if forecastL.PassingData != nil {
		updates.PassingEstimatedTime = forecastL.PassingData.EstimatedTime
		updates.PassingEstimatedWorkingTime = forecastL.PassingData.EstimatedWorkingTime
		updates.PassingMinimumEstimatedTime = forecastL.PassingData.EstimatedTimeMinimum
		updates.PassingEstimatedTimeIsUnknown = &forecastL.PassingData.EstimatedTimeUnknown
		updates.PassingIsDelayed = &forecastL.PassingData.Delayed
		updates.PassingActualTime = forecastL.PassingData.ActualTime
		updates.PassingActualTimeClass = forecastL.PassingData.ActualTimeClass
		updates.PassingDataSource = forecastL.PassingData.Source
		updates.PassingDataSourceSystem = forecastL.PassingData.SourceSystem
	}
	if forecastL.DepartureData != nil {
		updates.DepartureEstimatedTime = forecastL.DepartureData.EstimatedTime
		updates.DepartureEstimatedWorkingTime = forecastL.DepartureData.EstimatedWorkingTime
		updates.DepartureMinimumEstimatedTime = forecastL.DepartureData.EstimatedTimeMinimum
		updates.DepartureEstimatedTimeIsUnknown = &forecastL.DepartureData.EstimatedTimeUnknown
		updates.DepartureIsDelayed = &forecastL.DepartureData.Delayed
		updates.DepartureActualTime = forecastL.DepartureData.ActualTime
		updates.DepartureActualTimeClass = forecastL.DepartureData.ActualTimeClass
		updates.DepartureDataSource = forecastL.DepartureData.Source
		updates.DepartureDataSourceSystem = forecastL.DepartureData.SourceSystem
	}
	if forecastL.LateReason != nil {
		updates.LateReasonID = &forecastL.LateReason.ReasonID
		updates.LateReasonTIPLOC = forecastL.LateReason.TIPLOC
		updates.LateReasonIsNearLocation = &forecastL.LateReason.Near
	}
	if forecastL.DisruptionRisk != nil {
		updates.DisruptionRisk = &forecastL.DisruptionRisk.Effect
		if forecastL.DisruptionRisk.Reason != nil {
			updates.DisruptionRiskReasonID = &forecastL.DisruptionRisk.Reason.ReasonID
			updates.DisruptionRiskReasonTIPLOC = forecastL.DisruptionRisk.Reason.TIPLOC
			updates.DisruptionRiskReasonIsNearLocation = &forecastL.DisruptionRisk.Reason.Near
		}
	}
	updates.AffectedBy = forecastL.AffectedBy
	updates.Length = &forecastL.Length
	if forecastL.PlatformData != nil {
		updates.PlatformIsSuppressed = &forecastL.PlatformData.Suppressed
		updates.PlatformIsSuppressedByCIS = &forecastL.PlatformData.SuppressedByCIS
		updates.PlatformDataSource = &forecastL.PlatformData.Source
		updates.PlatformConfirmed = &forecastL.PlatformData.Confirmed
		updates.Platform = &forecastL.PlatformData.Platform
	}
	updates.IsSuppressed = &forecastL.Suppressed
	updates.DetachesFromFront = &forecastL.DetachesFromFront
	return updates
}

func (u UnitOfWork) updateOneScheduleLocationWithForecastUpdates(updates forecastScheduleLocationUpdates) error {
	_, err := u.tx.Exec(u.ctx, `
		UPDATE darwin.schedule_locations SET 
			arrival_estimated_time = @arrival_estimated_time
			,arrival_estimated_working_time = @arrival_estimated_working_time
			,arrival_minimum_estimated_time = @arrival_minimum_estimated_time
			,arrival_estimated_time_is_unknown = @arrival_estimated_time_is_unknown
			,arrival_is_delayed = @arrival_is_delayed
			,arrival_actual_time = @arrival_actual_time
			,arrival_actual_time_class = @arrival_actual_time_class
			,arrival_data_source = @arrival_data_source
			,arrival_data_source_system = @arrival_data_source_system
			,passing_estimated_time = @passing_estimated_time
			,passing_estimated_working_time = @passing_estimated_working_time
			,passing_minimum_estimated_time = @passing_minimum_estimated_time
			,passing_estimated_time_is_unknown = @passing_estimated_time_is_unknown
			,passing_is_delayed = @passing_is_delayed
			,passing_actual_time = @passing_actual_time
			,passing_actual_time_class = @passing_actual_time_class
			,passing_data_source = @passing_data_source
			,passing_data_source_system = @passing_data_source_system
			,departure_estimated_time = @departure_estimated_time
			,departure_estimated_working_time = @departure_estimated_working_time
			,departure_minimum_estimated_time = @departure_minimum_estimated_time
			,departure_estimated_time_is_unknown = @departure_estimated_time_is_unknown
			,departure_is_delayed = @departure_is_delayed
			,departure_actual_time = @departure_actual_time
			,departure_actual_time_class = @departure_actual_time_class
			,departure_data_source = @departure_data_source
			,departure_data_source_system = @departure_data_source_system
			,late_reason_id = @late_reason_id
			,late_reason_location_id = @late_reason_location_id
			,late_reason_is_near_location = @late_reason_is_near_location
			,disruption_risk = @disruption_risk
			,disruption_risk_reason_id = @disruption_risk_reason_id
			,disruption_risk_reason_location_id = @disruption_risk_reason_location_id
			,disruption_risk_reason_is_near_location = @disruption_risk_reason_is_near_location
			,affected_by = @affected_by
			,length = @length
			,platform_is_suppressed = @platform_is_suppressed
			,platform_is_suppressed_by_cis = @platform_is_suppressed_by_cis
			,platform_data_source = @platform_data_source
			,platform_confirmed = @platform_confirmed
			,platform = @platform
			,is_suppressed = @is_suppressed
			,detaches_from_front = @detaches_from_front
		WHERE schedule_id = @schedule_id AND sequence = @sequence
		`, pgx.StrictNamedArgs{
		"schedule_id":                             updates.ScheduleID,
		"sequence":                                updates.Sequence,
		"arrival_estimated_time":                  updates.ArrivalEstimatedTime,
		"arrival_estimated_working_time":          updates.ArrivalEstimatedWorkingTime,
		"arrival_minimum_estimated_time":          updates.ArrivalMinimumEstimatedTime,
		"arrival_estimated_time_is_unknown":       updates.ArrivalEstimatedTimeIsUnknown,
		"arrival_is_delayed":                      updates.ArrivalIsDelayed,
		"arrival_actual_time":                     updates.ArrivalActualTime,
		"arrival_actual_time_class":               updates.ArrivalActualTimeClass,
		"arrival_data_source":                     updates.ArrivalDataSource,
		"arrival_data_source_system":              updates.ArrivalDataSourceSystem,
		"passing_estimated_time":                  updates.PassingEstimatedTime,
		"passing_estimated_working_time":          updates.PassingEstimatedWorkingTime,
		"passing_minimum_estimated_time":          updates.PassingMinimumEstimatedTime,
		"passing_estimated_time_is_unknown":       updates.PassingEstimatedTimeIsUnknown,
		"passing_is_delayed":                      updates.PassingIsDelayed,
		"passing_actual_time":                     updates.PassingActualTime,
		"passing_actual_time_class":               updates.PassingActualTimeClass,
		"passing_data_source":                     updates.PassingDataSource,
		"passing_data_source_system":              updates.PassingDataSourceSystem,
		"departure_estimated_time":                updates.DepartureEstimatedTime,
		"departure_estimated_working_time":        updates.DepartureEstimatedWorkingTime,
		"departure_minimum_estimated_time":        updates.DepartureMinimumEstimatedTime,
		"departure_estimated_time_is_unknown":     updates.DepartureEstimatedTimeIsUnknown,
		"departure_is_delayed":                    updates.DepartureIsDelayed,
		"departure_actual_time":                   updates.DepartureActualTime,
		"departure_actual_time_class":             updates.DepartureActualTimeClass,
		"departure_data_source":                   updates.DepartureDataSource,
		"departure_data_source_system":            updates.DepartureDataSourceSystem,
		"late_reason_id":                          updates.LateReasonID,
		"late_reason_location_id":                 updates.LateReasonTIPLOC,
		"late_reason_is_near_location":            updates.LateReasonIsNearLocation,
		"disruption_risk":                         updates.DisruptionRisk,
		"disruption_risk_reason_id":               updates.DisruptionRiskReasonID,
		"disruption_risk_reason_location_id":      updates.DisruptionRiskReasonTIPLOC,
		"disruption_risk_reason_is_near_location": updates.DisruptionRiskReasonIsNearLocation,
		"affected_by":                             updates.AffectedBy,
		"length":                                  updates.Length,
		"platform_is_suppressed":                  updates.PlatformIsSuppressed,
		"platform_is_suppressed_by_cis":           updates.PlatformIsSuppressedByCIS,
		"platform_data_source":                    updates.PlatformDataSource,
		"platform_confirmed":                      updates.PlatformConfirmed,
		"platform":                                updates.Platform,
		"is_suppressed":                           updates.IsSuppressed,
		"detaches_from_front":                     updates.DetachesFromFront,
	})
	return err
}
