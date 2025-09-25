package repository

type AssociationRow struct {
	TIPLOC                    string  `db:"tiploc"`
	Category                  string  `db:"category"`
	Cancelled                 bool    `db:"is_cancelled"`
	Deleted                   bool    `db:"is_deleted"`
	MainServiceID             string  `db:"main_service_id"`
	MainWorkingArrivalTime    *string `db:"main_working_arrival_time"`
	MainWorkingDepartureTime  *string `db:"main_working_departure_time"`
	MainWorkingPassingTime    *string `db:"main_working_passing_time"`
	MainPublicArrivalTime     *string `db:"main_public_arrival_time"`
	MainPublicDepartureTime   *string `db:"main_public_departure_time"`
	AssocServiceRID           string  `db:"assoc_service_id"`
	AssocWorkingArrivalTime   *string `db:"assoc_working_arrival_time"`
	AssocWorkingDepartureTime *string `db:"assoc_working_departure_time"`
	AssocWorkingPassingTime   *string `db:"assoc_working_passing_time"`
	AssocPublicArrivalTime    *string `db:"assoc_public_arrival_time"`
	AssocPublicDepartureTime  *string `db:"assoc_public_departure_time"`
}
