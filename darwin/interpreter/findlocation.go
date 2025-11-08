package interpreter

import (
	"strings"

	"github.com/headblockhead/railreader/darwin/unmarshaller"
	"github.com/jackc/pgx/v5"
)

// findLocationSequence finds the sequence number of a location in a schedule by the ID of the schedule it belongs to, the ID of the location the entry is at, and a set of times that the location must match when specified.
// This is needed because sometimes schedules can have multiple stops at the same location but at different times - for example with circular routes.
// PublicArrivalTime and PublicDepartureTime are ignored when matching times.
func (u *UnitOfWork) findLocationSequence(scheduleID string, locationID string, times unmarshaller.LocationTimeIdentifiers) (int, error) {
	var b strings.Builder

	_, err := b.WriteString(`SELECT sequence FROM darwin.schedule_locations WHERE schedule_id = @rid AND location_id = @location_id`)
	if err != nil {
		return 0, err
	}
	args := pgx.StrictNamedArgs{
		"schedule_id": scheduleID,
		"location_id": locationID,
	}

	// For reasons that I do not know, LocationTimeIdentifiers will sometimes specify PublicArrivalTime and PublicDepartureTimes, even though the location they refer to does not have them.
	// Because of this, I don't check for a match of PublicArrivalTime and PublicDepartureTime here, and instead only check Working times.
	// This may only be the case for when loading from a timetable, but I haven't checked yet.
	// TODO: check if locationtimeids only [set pta+ptd when the location doesn't have them] when loading from a timetable.

	if times.WorkingArrivalTime != nil {
		if _, err = b.WriteString(" AND working_arrival_time = @working_arrival_time"); err != nil {
			return 0, err
		}
		args["working_arrival_time"] = *times.WorkingArrivalTime
	}
	if times.WorkingPassingTime != nil {
		if _, err = b.WriteString(" AND working_passing_time = @working_passing_time"); err != nil {
			return 0, err
		}
		args["working_passing_time"] = *times.WorkingPassingTime
	}
	if times.WorkingDepartureTime != nil {
		if _, err := b.WriteString(" AND working_departure_time = @working_departure_time"); err != nil {
			return 0, err
		}
		args["working_departure_time"] = *times.WorkingDepartureTime
	}
	if _, err := b.WriteRune(';'); err != nil {
		return 0, err
	}

	var sequence int
	row := u.tx.QueryRow(u.ctx, b.String(), args)
	if err := row.Scan(&sequence); err != nil {
		return 0, err
	}

	return sequence, nil
}
