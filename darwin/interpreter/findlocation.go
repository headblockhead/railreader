package interpreter

import (
	"database/sql"
	"fmt"

	"github.com/headblockhead/railreader/darwin/unmarshaller"
)

// findLocationSequence finds the sequence number of a location in a schedule by the ID of the schedule it belongs to, the ID of the location the entry is at, and a set of times that the location must match when specified.
// This is needed because sometimes schedules can have multiple stops at the same location but at different times - for example with circular routes.
// PublicArrivalTime and PublicDepartureTime are ignored when matching times.
func (u *UnitOfWork) findLocationSequence(scheduleID string, locationID string, times unmarshaller.LocationTimeIdentifiers) (int, error) {
	statement := `SELECT sequence FROM schedule_locations WHERE schedule_id = :rid AND location_id = :location_id`
	args := []sql.NamedArg{
		sql.Named("rid", scheduleID),
		sql.Named("location_id", locationID),
	}

	// For reasons that I do not know, [associations, TODO: possibly more?] will sometimes specify PublicArrivalTime and PublicDepartureTimes, even though the location they refer to does not have them.
	// Because of this, I don't check for a match of PublicArrivalTime and PublicDepartureTime here, and instead only check Working times.

	if times.WorkingArrivalTime != nil {
		statement += " AND working_arrival_time = :working_arrival_time"
		args = append(args, sql.Named("working_arrival_time", *times.WorkingArrivalTime))
	}
	if times.WorkingPassingTime != nil {
		statement += " AND working_passing_time = :working_passing_time"
		args = append(args, sql.Named("working_passing_time", *times.WorkingPassingTime))
	}
	if times.WorkingDepartureTime != nil {
		statement += " AND working_departure_time = :working_departure_time"
		args = append(args, sql.Named("working_departure_time", *times.WorkingDepartureTime))
	}
	statement += ";"

	var sequence int
	row := u.tx.QueryRow(statement)
	if err := row.Scan(&sequence); err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("findLocationSequence (scheduleID %s) (locationID %s) (workingArrivalTime %v) (workingPassingTime %v) (workingDepartureTime %v): no matching location found", scheduleID, locationID, times.WorkingArrivalTime, times.WorkingPassingTime, times.WorkingDepartureTime)
		}
		return 0, fmt.Errorf("findLocationSequence (scheduleID %s) (locationID %s) (workingArrivalTime %v) (workingPassingTime %v) (workingDepartureTime %v): %v", scheduleID, locationID, times.WorkingArrivalTime, times.WorkingPassingTime, times.WorkingDepartureTime, err)
	}

	return sequence, nil
}
