package darwin

type ServiceType string

func (tt ServiceType) String() string {
	return map[ServiceType]string{
		PassengerAndParcelTrain:     "Train",
		Bus:                         "Bus",
		Ship:                        "Ship",
		Trip:                        "Trip",
		Freight:                     "Freight",
		PassengerAndParcelShortTerm: "Train",
		BusShortTerm:                "Bus",
		ShipShortTerm:               "Ship",
		TripShortTerm:               "Trip",
		FreightShortTerm:            "Freight",
	}[tt]
}

const (
	PassengerAndParcelTrain     ServiceType = "P"
	Bus                         ServiceType = "B"
	Ship                        ServiceType = "S"
	Trip                        ServiceType = "T"
	Freight                     ServiceType = "F"
	PassengerAndParcelShortTerm ServiceType = "1"
	BusShortTerm                ServiceType = "5"
	ShipShortTerm               ServiceType = "4"
	TripShortTerm               ServiceType = "3"
	FreightShortTerm            ServiceType = "2"
)

type ActivityCode string

// TODO
/*const (*/
/*None                                     ActivityCode = "  "*/
/*StopsToDetatchVehicles                   ActivityCode = "-D"*/
/*StopsToAttachAndDetachVehicles           ActivityCode = "-T"*/
/*StopsToAttachVehicles                    ActivityCode = "-U"*/
/*StopsOrShuntsForOtherTrainsToPass        ActivityCode = "A "*/
/*StopsToAttachOrDetachAssistingLocomotive ActivityCode = "AE"*/
/*)*/

type DisruptionReason string

// TODO
const ()

// TIPLOC is a code representing a location.
// TIPLOCs can be a station, junction, or any other relevant location.
type TIPLOC string

// TODO: map TIPLOCs to human-readable names using dataset.

// TrainTime is a time in HH:MM or HH:MM:SS format, representing a time of day.
type TrainTime string
