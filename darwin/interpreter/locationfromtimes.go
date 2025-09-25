package interpreter

import (
	"errors"
	"time"

	"github.com/headblockhead/railreader"
	"github.com/headblockhead/railreader/darwin/repository"
	"github.com/headblockhead/railreader/darwin/unmarshaller"
)

func (u UnitOfWork) locationFromTIPLOCAndTimes(scheduleID string, tiploc railreader.TimingPointLocationCode, times unmarshaller.LocationTimeIdentifiers) (row repository.ScheduleLocationRow, err error) {
	possibleLocations, err := u.scheduleLocationRepository.SelectByScheduleIDAndLocationID(scheduleID, string(tiploc))
	if err != nil {
		return repository.ScheduleLocationRow{}, err
	}
	if len(possibleLocations) == 0 {
		return repository.ScheduleLocationRow{}, errors.New("no locations found for schedule")
	}
	if len(possibleLocations) == 1 {
		return possibleLocations[0], nil
	}
	for _, loc := range possibleLocations {
		points := 0
		possiblePoints := 0
		if times.WorkingArrivalTime != nil {
			possiblePoints++
			// Get the time, ignore the date
			wta, err := trainTimeToTime(time.Time{}, *times.WorkingArrivalTime, time.Time{})
			if err != nil {
				return repository.ScheduleLocationRow{}, err
			}
			if loc.WorkingArrivalTime != nil {
				if wta.Hour() == loc.WorkingArrivalTime.Hour() && wta.Minute() == loc.WorkingArrivalTime.Minute() {
					points++
				}
			}
		}
		if times.WorkingDepartureTime != nil {
			possiblePoints++
			wtd, err := trainTimeToTime(time.Time{}, *times.WorkingDepartureTime, time.Time{})
			if err != nil {
				return repository.ScheduleLocationRow{}, err
			}
			if loc.WorkingDepartureTime != nil {
				if wtd.Hour() == loc.WorkingDepartureTime.Hour() && wtd.Minute() == loc.WorkingDepartureTime.Minute() {
					points++
				}
			}
		}
		if times.WorkingPassingTime != nil {
			possiblePoints++
			wtp, err := trainTimeToTime(time.Time{}, *times.WorkingPassingTime, time.Time{})
			if err != nil {
				return repository.ScheduleLocationRow{}, err
			}
			if loc.WorkingPassingTime != nil {
				if wtp.Hour() == loc.WorkingPassingTime.Hour() && wtp.Minute() == loc.WorkingPassingTime.Minute() {
					points++
				}
			}
		}
		if times.PublicArrivalTime != nil {
			possiblePoints++
			pta, err := trainTimeToTime(time.Time{}, *times.PublicArrivalTime, time.Time{})
			if err != nil {
				return repository.ScheduleLocationRow{}, err
			}
			if loc.PublicArrivalTime != nil {
				if pta.Hour() == loc.PublicArrivalTime.Hour() && pta.Minute() == loc.PublicArrivalTime.Minute() {
					points++
				}
			}
		}
		if times.PublicDepartureTime != nil {
			possiblePoints++
			ptd, err := trainTimeToTime(time.Time{}, *times.PublicDepartureTime, time.Time{})
			if err != nil {
				return repository.ScheduleLocationRow{}, err
			}
			if loc.PublicDepartureTime != nil {
				if ptd.Hour() == loc.PublicDepartureTime.Hour() && ptd.Minute() == loc.PublicDepartureTime.Minute() {
					points++
				}
			}
		}
		if points == possiblePoints && points > 0 {
			return loc, nil
		}
	}
	return repository.ScheduleLocationRow{}, errors.New("no matching location found for schedule")
}
