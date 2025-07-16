package darwin

// ServiceType represents what mode of transport a service is.
type ServiceType string

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

var ServiceTypeToString = map[ServiceType]string{
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
}

func (tt ServiceType) String() string {
	return ServiceTypeToString[tt]
}

type ActivityCode string

// Some of the meaning of these codes are unknown/unclear.
const (
	None                                                ActivityCode = "  "
	StopsOrShuntsForOtherTrainsToPass                   ActivityCode = "A "
	StopsToAttachOrDetachAssistingLocomotive            ActivityCode = "AE"
	ShowsAsXOnArrival                                   ActivityCode = "AX"
	StopsForBankingLocomotive                           ActivityCode = "BL"
	StopsToChangeTrainsmen                              ActivityCode = "C "
	StopsToSetDownPassengers                            ActivityCode = "D "
	StopsToDetatchVehicles                              ActivityCode = "-D"
	StopsForExamination                                 ActivityCode = "E "
	NationalRailTimetableDataToAdd                      ActivityCode = "G "
	NotionalActivityToPreventWTTTimingColumnsMerge      ActivityCode = "H "
	NotionalActivityToPreventWTTTimingColumnsMergeTwice ActivityCode = "HH"
	PassengerCountPoint                                 ActivityCode = "K "
	TicketCollectionAndExaminationPoint                 ActivityCode = "KC"
	TicketExaminationPoint                              ActivityCode = "KE"
	TicketExaminationPointFirstClassOnly                ActivityCode = "KF"
	TicketExaminationPointSelective                     ActivityCode = "KS"
	StopsToChangeLocomotives                            ActivityCode = "L "
	StopNotAdvertised                                   ActivityCode = "N "
	StopsForOtherOperatingReasons                       ActivityCode = "OP"
	TrainLocomotiveOnRear                               ActivityCode = "OR"
	PropellingBetweenPointsShown                        ActivityCode = "PR"
	StopsWhenRequired                                   ActivityCode = "R "
	ReversingMovementOrDriverChangesEnds                ActivityCode = "RM"
	StopsForLocomotiveToRunRound                        ActivityCode = "RR"
	StopsForRailwayPersonellOnly                        ActivityCode = "S "
	StopsToTakeUpAndSetDownPassengers                   ActivityCode = "T "
	StopsToAttachAndDetachVehicles                      ActivityCode = "-T"
	TrainBegins                                         ActivityCode = "TB"
	TrainFinishes                                       ActivityCode = "TF"
	DetailConsistForTOPSDirect                          ActivityCode = "TS"
	StopsOrPassesForTabletStaffOrToken                  ActivityCode = "TW"
	StopsToTakeUpPassengers                             ActivityCode = "U "
	StopsToAttachVehicles                               ActivityCode = "-U"
	StopsForWateringOfCoaches                           ActivityCode = "W "
	PassesAnotherTrainAtCrossingPointOnSingleLine       ActivityCode = "X "
)

type TrainCategory string

const (
	// O - Ordinary
	UndergroundOrMetro    TrainCategory = "OL"
	UnadvertisedPassenger TrainCategory = "OU"
	Passenger             TrainCategory = "OO"
	Staff                 TrainCategory = "OS"
	Mixed                 TrainCategory = "OW"
	// X - Express
	ChannelTunnel TrainCategory = "XC"

	// TODO
)

type DisruptionReason string

// TODO

// TIPLOC is a code representing a location.
// TIPLOCs can be a station, junction, or any other relevant location.
type TIPLOC string

// TODO: map TIPLOCs to human-readable names using dataset.

// TrainTime is a time in HH:MM or HH:MM:SS format, representing a time of day.
type TrainTime string
