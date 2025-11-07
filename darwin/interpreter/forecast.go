package interpreter

import "github.com/headblockhead/railreader/darwin/unmarshaller"

func (u UnitOfWork) interpretForecast(forecast unmarshaller.ForecastTime) error {
	record, err := u.forecastToRecord(forecast)
	if err != nil {
		return err
	}
	return u.updateScheduleWithForecastRecord(record)
}

type ForecastRecord struct {
	ScheduleID string
	Sequence   int

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
	DepartureEstimatedTimeIsUnknown bool
	DepartureIsDelayed              bool
	DepartureActualTime             *string
	DepartureActualTimeClass        *string
	DepartureDataSource             *string
	DepartureDataSourceSystem       *string

	LateReasonID             *int
	LateReasonLocationID     *string
	LateReasonIsNearLocation *bool

	DisruptionRisk                     *string
	DisruptionRiskReasonID             *int
	DisruptionRiskReasonLocationID     *string
	DisruptionRiskReasonIsNearLocation *bool

	AffectedBy             *string
	Length                 int
	PlatformSupressed      bool
	PlatformSupressedByCIS *bool
	PlatformDataSource     *rune
	PlatformConfirmed      *bool
	Platform               *string
}

func (u UnitOfWork) forecastToRecord(forecast unmarshaller.ForecastTime) (ForecastRecord, error) {
	var record ForecastRecord
	return record, nil
}

func (u UnitOfWork) updateScheduleWithForecastRecord(record ForecastRecord) error {
	return nil
}
