package railreader

// ServiceType represents what mode of transport a service is.
type ServiceType string

const (
	ServicePassengerOrParcelTrain     ServiceType = "P"
	ServiceBus                        ServiceType = "B"
	ServiceShip                       ServiceType = "S"
	ServiceTrip                       ServiceType = "T"
	ServiceFreight                    ServiceType = "F"
	ServicePassengerOrParcelShortTerm ServiceType = "1"
	ServiceBusShortTerm               ServiceType = "5"
	ServiceShipShortTerm              ServiceType = "4"
	ServiceTripShortTerm              ServiceType = "3"
	ServiceFreightShortTerm           ServiceType = "2"
)

var ServiceTypeStrings = map[ServiceType]string{
	ServicePassengerOrParcelTrain:     "Train",
	ServiceBus:                        "Bus",
	ServiceShip:                       "Ship",
	ServiceTrip:                       "Trip",    // not included in Darwin
	ServiceFreight:                    "Freight", // not included in Darwin
	ServicePassengerOrParcelShortTerm: "Train",
	ServiceBusShortTerm:               "Bus",
	ServiceShipShortTerm:              "Ship",
	ServiceTripShortTerm:              "Trip",    // not included in Darwin
	ServiceFreightShortTerm:           "Freight", // not included in Darwin
}

func (tt ServiceType) String() string {
	if str, ok := ServiceTypeStrings[tt]; ok {
		return str
	}
	return "Unknown"
}

type ActivityCode string

// Some of the meaning of these codes are unknown/unclear, and are commented out.
const (
	ActivityNone ActivityCode = "  "

	ActivityTrainBegins   ActivityCode = "TB"
	ActivityTrainFinishes ActivityCode = "TF"

	ActivityStopsToTakeUpAndSetDownPassengers ActivityCode = "T "
	ActivityStopsToTakeUpPassengers           ActivityCode = "U "
	ActivityStopsToSetDownPassengers          ActivityCode = "D "

	ActivityStopsOrShuntsForOtherTrainsToPass        ActivityCode = "A "
	ActivityStopsToAttachOrDetachAssistingLocomotive ActivityCode = "AE"
	//ActivityShowsAsXOnArrival                                   ActivityCode = "AX"
	ActivityStopsForBankingLocomotive ActivityCode = "BL"
	ActivityStopsToChangeTrainCrew    ActivityCode = "C "
	ActivityStopsToDetatchVehicles    ActivityCode = "-D"
	ActivityStopsForExamination       ActivityCode = "E "
	//ActivityNationalRailTimetableDataToAdd                      ActivityCode = "G "
	//ActivityNotionalActivityToPreventWTTTimingColumnsMerge      ActivityCode = "H "
	//ActivityNotionalActivityToPreventWTTTimingColumnsMergeTwice ActivityCode = "HH"
	ActivityPassengerCountPoint                  ActivityCode = "K "
	ActivityTicketCollectionAndExaminationPoint  ActivityCode = "KC"
	ActivityTicketExaminationPoint               ActivityCode = "KE"
	ActivityTicketExaminationPointFirstClassOnly ActivityCode = "KF"
	ActivityTicketExaminationPointSelective      ActivityCode = "KS"
	ActivityStopsToChangeLocomotives             ActivityCode = "L "
	ActivityStopNotAdvertised                    ActivityCode = "N "
	ActivityStopsForOtherOperatingReasons        ActivityCode = "OP"
	//ActivityTrainLocomotiveOnRear                         ActivityCode = "OR"
	ActivityPropellingBetweenPointsShown               ActivityCode = "PR"
	ActivityStopsWhenRequired                          ActivityCode = "R "
	ActivityStopsForReversingMovementOrDriverEndChange ActivityCode = "RM"
	ActivityStopsForLocomotiveToRunRound               ActivityCode = "RR"
	ActivityStopsForRailwayPersonellOnly               ActivityCode = "S "
	ActivityStopsToAttachAndDetachVehicles             ActivityCode = "-T"
	//ActivityDetailConsistForTOPSDirect                    ActivityCode = "TS"
	ActivityStopsOrPassesForTabletStaffOrToken            ActivityCode = "TW"
	ActivityStopsToAttachVehicles                         ActivityCode = "-U"
	ActivityStopsForWateringOfCoaches                     ActivityCode = "W "
	ActivityPassesAnotherTrainAtCrossingPointOnSingleLine ActivityCode = "X "
)

var ActivityCodeStrings = map[ActivityCode]string{
	ActivityTrainBegins:                                   "begins",
	ActivityTrainFinishes:                                 "finishes",
	ActivityStopsToTakeUpAndSetDownPassengers:             "stops",
	ActivityStopsToTakeUpPassengers:                       "takes up passengers",
	ActivityStopsToSetDownPassengers:                      "sets down passengers",
	ActivityStopsOrShuntsForOtherTrainsToPass:             "stops/shunts for other trains to pass",
	ActivityStopsToAttachOrDetachAssistingLocomotive:      "stops to attach/detach an assisting locomotive",
	ActivityStopsForBankingLocomotive:                     "stops for a banking locomotive",
	ActivityStopsToChangeTrainCrew:                        "stops to change train crew",
	ActivityStopsToDetatchVehicles:                        "stops to detach vehicles",
	ActivityStopsForExamination:                           "stops for examination",
	ActivityPassengerCountPoint:                           "passengers counted",
	ActivityTicketCollectionAndExaminationPoint:           "tickets collected and examined",
	ActivityTicketExaminationPoint:                        "tickets examined",
	ActivityTicketExaminationPointFirstClassOnly:          "tickets for first class examined",
	ActivityTicketExaminationPointSelective:               "tickets possibly examined",
	ActivityStopsToChangeLocomotives:                      "stops to change locomotives",
	ActivityStopNotAdvertised:                             "stops (unadvertised)",
	ActivityStopsForOtherOperatingReasons:                 "stops for operating reasons",
	ActivityPropellingBetweenPointsShown:                  "propells between points shown",
	ActivityStopsWhenRequired:                             "stops when required",
	ActivityStopsForReversingMovementOrDriverEndChange:    "stops to reverse",
	ActivityStopsForLocomotiveToRunRound:                  "stops for locomotive to run round",
	ActivityStopsForRailwayPersonellOnly:                  "stops (railway personnel only)",
	ActivityStopsToAttachAndDetachVehicles:                "stops to attach/detach vehicles",
	ActivityStopsOrPassesForTabletStaffOrToken:            "stops/passes for tablet staff/token",
	ActivityStopsToAttachVehicles:                         "stops to attach vehicles",
	ActivityStopsForWateringOfCoaches:                     "stops for watering of coaches",
	ActivityPassesAnotherTrainAtCrossingPointOnSingleLine: "passes another train at crossing point on a single line",
}

func (ac ActivityCode) String() string {
	if str, ok := ActivityCodeStrings[ac]; ok {
		return str
	}
	return "does an unknown activity"
}

// ServiceCategory describes the function of a service.
// The String() form can be prepended to the ServiceType when describing a service to give additional information about it.
type ServiceCategory string

const (
	// O - Ordinary
	CategoryUndergroundOrMetro    ServiceCategory = "OL"
	CategoryUnadvertisedPassenger ServiceCategory = "OU"
	CategoryPassenger             ServiceCategory = "OO"
	CategoryStaff                 ServiceCategory = "OS"
	CategoryMixed                 ServiceCategory = "OW"
	// X - Express
	CategoryChannelTunnel       ServiceCategory = "XC"
	CategorySleeper             ServiceCategory = "XD"
	CategoryInternational       ServiceCategory = "XI"
	CategoryMotorail            ServiceCategory = "XR"
	CategoryUnadvertisedExpress ServiceCategory = "XU"
	CategoryExpress             ServiceCategory = "XX"
	CategorySleeperDomestic     ServiceCategory = "XZ"
	// B - Bus
	CategoryBusReplacement ServiceCategory = "BR"
	CategoryBusService     ServiceCategory = "BS"
	// S - Ship
	CategoryShip ServiceCategory = "SS"
	// E - Empty Coaching Stock
	CategoryEmptyCoachingStock                   ServiceCategory = "EE"
	CategoryEmptyCoachingStockUndergroundOrMetro ServiceCategory = "EL"
	CategoryEmptyCoachingStockOrStaff            ServiceCategory = "ES"
	// Parcels and Postal trains
	CategoryPostal                                 ServiceCategory = "JJ"
	CategoryPostOfficeControlledParcels            ServiceCategory = "PM"
	CategoryParcels                                ServiceCategory = "PP"
	CategoryEmptyNonPassengerCarryingCoachingStock ServiceCategory = "PV"
	// D - Departmental trains
	CategoryDepartmental                          ServiceCategory = "DD"
	CategoryCivilEngineering                      ServiceCategory = "DH"
	CategoryMechanicalOrElectricalEngineering     ServiceCategory = "DI"
	CategoryDepartmentalStores                    ServiceCategory = "DQ"
	CategoryTest                                  ServiceCategory = "DT"
	CategorySignalOrTelecommunicationsEngineering ServiceCategory = "DY"
	// Z - Light locomotives
	CategoryLocomotiveOrBrakeVan ServiceCategory = "ZB"
	CategoryLightLocomotive      ServiceCategory = "ZZ"
	// Railfreight distribution
	CategoryRfDAutomotiveComponents  ServiceCategory = "J2"
	CategoryRfDAutomotives           ServiceCategory = "H2"
	CategoryRfDEdibleProducts        ServiceCategory = "J3"
	CategoryRfDIndustrialMinerals    ServiceCategory = "J4"
	CategoryRfDChemicals             ServiceCategory = "J5"
	CategoryRfDBuildingMaterials     ServiceCategory = "J6"
	CategoryRfDGeneralMerchandise    ServiceCategory = "J8"
	CategoryRfDEuropean              ServiceCategory = "H8"
	CategoryRfDFreightlinerContracts ServiceCategory = "J9"
	CategoryRfDFreightlinerOther     ServiceCategory = "H9"
	// Trainload Freight
	CategoryTLFCoalDistribution          ServiceCategory = "A0"
	CategoryTLFCoalForEletricity         ServiceCategory = "E0"
	CategoryTLFCoalOrNuclear             ServiceCategory = "B0"
	CategoryTLFMetals                    ServiceCategory = "B1"
	CategoryTLFAggregates                ServiceCategory = "B4"
	CategoryTLFDomesticOrIndustrialWaste ServiceCategory = "B5"
	CategoryTLFBuildingMaterials         ServiceCategory = "B6"
	CategoryTLFPetroleumProducts         ServiceCategory = "B7"
	// Railfreight distribution through the Channel Tunnel
	CategoryRfDChannelMixedBusiness    ServiceCategory = "H0"
	CategoryRfDChannelIntermodal       ServiceCategory = "H1"
	CategoryRfDChannelAutomotive       ServiceCategory = "H3"
	CategoryRfDChannelContractServices ServiceCategory = "H4"
	CategoryRfDChannelHaulmark         ServiceCategory = "H5"
	CategoryRfDChannelJointVenture     ServiceCategory = "H6"
)

var ServiceCategoryStrings = map[ServiceCategory]string{
	// O - Ordinary
	CategoryUndergroundOrMetro:    "Underground/Metro",
	CategoryUnadvertisedPassenger: "(unadvertised) Passenger",
	CategoryPassenger:             "Passenger",
	CategoryStaff:                 "Staff",
	CategoryMixed:                 "Mixed",
	// X - Express
	CategoryChannelTunnel:       "Channel Tunnel",
	CategorySleeper:             "Sleeper",
	CategoryInternational:       "International",
	CategoryMotorail:            "Motorail",
	CategoryUnadvertisedExpress: "(unadvertised) Express",
	CategoryExpress:             "Express",
	CategorySleeperDomestic:     "Sleeper",
	// B - Bus
	CategoryBusReplacement: "Rail Replacement",
	CategoryBusService:     "",
	// S - Ship
	CategoryShip: "Ship",
	// E - Empty Coaching Stock
	CategoryEmptyCoachingStock:                   "Empty",
	CategoryEmptyCoachingStockUndergroundOrMetro: "Empty Underground/Metro",
	CategoryEmptyCoachingStockOrStaff:            "Empty/Staff",
	// Parcels and Postal trains
	CategoryPostal:                                 "Postal",
	CategoryPostOfficeControlledParcels:            "Post Office Controlled Parcel",
	CategoryParcels:                                "Parcel",
	CategoryEmptyNonPassengerCarryingCoachingStock: "Empty Non-Passenger",
	// D - Departmental trains
	CategoryDepartmental:                          "Departmental",
	CategoryCivilEngineering:                      "Civil Engineering",
	CategoryMechanicalOrElectricalEngineering:     "Mechanical/Electrical Engineering",
	CategoryDepartmentalStores:                    "Departmental Store",
	CategoryTest:                                  "Test",
	CategorySignalOrTelecommunicationsEngineering: "Signal/Telecommunications Engineering",
	// Z - Light locomotives
	CategoryLocomotiveOrBrakeVan: "Brake Van/Locomotive",
	CategoryLightLocomotive:      "Light Locomotive",
	// Railfreight distribution
	CategoryRfDAutomotiveComponents:  "Automotive Components Freight",
	CategoryRfDAutomotives:           "Automotive Freight",
	CategoryRfDEdibleProducts:        "Edible Product Freight",
	CategoryRfDIndustrialMinerals:    "Industrial Mineral Freight",
	CategoryRfDChemicals:             "Chemical Freight",
	CategoryRfDBuildingMaterials:     "Building Material Freight",
	CategoryRfDGeneralMerchandise:    "General Merchandise Freight",
	CategoryRfDEuropean:              "European Freight",
	CategoryRfDFreightlinerContracts: "Contract Freightliner",
	CategoryRfDFreightlinerOther:     "Other Freightliner",
	// Trainload Freight
	CategoryTLFCoalDistribution:          "Coal Distribution Trainload Freight",
	CategoryTLFCoalForEletricity:         "Coal (for Electricity) Trainload Freight",
	CategoryTLFCoalOrNuclear:             "Coal/Nuclear Trainload Freight",
	CategoryTLFMetals:                    "Metal Trainload Freight",
	CategoryTLFAggregates:                "Aggregate Trainload Freight",
	CategoryTLFDomesticOrIndustrialWaste: "Domestic/Industrial Waste Trainload Freight",
	CategoryTLFBuildingMaterials:         "Building Material Trainload Freight",
	CategoryTLFPetroleumProducts:         "Petroleum Product Trainload Freight",
	// Railfreight distribution through the Channel Tunnel
	CategoryRfDChannelMixedBusiness:    "Mixed Business Channel Tunnel Freight",
	CategoryRfDChannelIntermodal:       "Intermodal Channel Tunnel Freight",
	CategoryRfDChannelAutomotive:       "Automotive Channel Tunnel Freight",
	CategoryRfDChannelContractServices: "Contract Services Channel Tunnel Freight",
	CategoryRfDChannelHaulmark:         "Haulmark Channel Tunnel Freight",
	CategoryRfDChannelJointVenture:     "Joint Venture Channel Tunnel Freight",
}

func (sc ServiceCategory) String() string {
	if str, ok := ServiceCategoryStrings[sc]; ok {
		return str
	}
	return "Unknown"
}

// TimingPointLocationCode (TIPLOC) is a code representing a location.
// TIPLOCs can be a passenger station, junction, or any other relevant location.
type TimingPointLocationCode string

type AssociationCategory string

const (
	// AssociationJoin indicates that two services join together into a single train.
	AssociationJoin AssociationCategory = "JJ"
	// AssociationDivide indicates a single train divides into two services, and one of them terminates.
	AssociationDivide AssociationCategory = "VV"
	// AssociationLink indicates two services are linked together into a single service.
	// For example, a train that terminates halfway through its schedule, and a bus replacement that continues the service.
	// Services are not necessarily linked at their termini, and links may create branching paths.
	// This is different from a join/divide as links may involve a change of service type.
	// In Darwin, passengers transfer from the MainService to the AssociatedService.
	AssociationLink AssociationCategory = "LK"
	// AssociationNext indicates the next service to be run using the same rolling stock.
	AssociationNext AssociationCategory = "NP"
)

// TrainDescriber is a two-letter code.
type TrainDescriber string

// TrainDescriberBerth is a four-character area of track dictated usually by a signal. Verths are not unique between TrainDescribers.
type TrainDescriberBerth string
