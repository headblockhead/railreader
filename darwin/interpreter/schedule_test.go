package interpreter

import (
	"log/slog"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/headblockhead/railreader"
	"github.com/headblockhead/railreader/darwin/database"
	"github.com/headblockhead/railreader/darwin/unmarshaller"
)

type testingScheduleRepository struct {
	// Schedules is a map of ScheduleID to Schedule, used to fake a database.
	Schedules map[string]database.Schedule
}

func (sr *testingScheduleRepository) Insert(schedule database.Schedule) error {
	sr.Schedules[schedule.ScheduleID] = schedule
	return nil
}

func TestInterpretSchedule(t *testing.T) {
	sr := &testingScheduleRepository{Schedules: make(map[string]database.Schedule)}
	log := slog.New(slog.NewTextHandler(nil, nil))
	location, err := time.LoadLocation("Europe/London")
	if err != nil {
		t.Fatalf("Failed to load location: %v", err)
	}
	testCases := []struct {
		// Inputs
		Schedule  unmarshaller.Schedule
		MessageID string
		// Expected outputs
		ExpectedDatabaseSchedule database.Schedule
	}{
		// Minimum valid schedule
		{
			Schedule: unmarshaller.Schedule{
				TrainIdentifiers: unmarshaller.TrainIdentifiers{
					RID:                "012345678901234",
					UID:                "A00001",
					ScheduledStartDate: "2025-08-23",
				},
				Headcode:         "2C04",
				TOC:              "NT",
				Service:          railreader.ServicePassengerOrParcelTrain,
				Category:         railreader.CategoryPassenger,
				PassengerService: true,
				Active:           true,
				Locations: []unmarshaller.ScheduleLocation{
					{
						Type: unmarshaller.LocationTypeOrigin,
						Origin: &unmarshaller.OriginLocation{
							LocationBase: unmarshaller.LocationBase{
								TIPLOC: "ABCD",
							},
							WorkingDepartureTime: "00:04",
						},
					},
					{
						Type: unmarshaller.LocationTypeDestination,
						Destination: &unmarshaller.DestinationLocation{
							LocationBase: unmarshaller.LocationBase{
								TIPLOC: "EFGH",
							},
							WorkingArrivalTime: "00:16",
						},
					},
				},
			},
			MessageID: "message1",
			ExpectedDatabaseSchedule: database.Schedule{
				ScheduleID:              "012345678901234",
				MessageID:               "message1",
				UID:                     "A00001",
				ScheduledStartDate:      time.Date(2025, 8, 23, 0, 0, 0, 0, location),
				Headcode:                "2C04",
				TrainOperatingCompanyID: "NT",
				Service:                 string(railreader.ServicePassengerOrParcelTrain),
				Category:                string(railreader.CategoryPassenger),
				Active:                  true,
				Locations: []database.ScheduleLocation{
					{
						Sequence:             0,
						LocationID:           "ABCD",
						Type:                 string(unmarshaller.LocationTypeOrigin),
						WorkingDepartureTime: pointerTo(time.Date(2025, 8, 23, 0, 4, 0, 0, location)),
					},
					{
						Sequence:           1,
						LocationID:         "EFGH",
						Type:               string(unmarshaller.LocationTypeDestination),
						WorkingArrivalTime: pointerTo(time.Date(2025, 8, 23, 0, 16, 0, 0, location)),
					},
				},
			},
		},
		// Schedule with all fields populated
		{
			Schedule: unmarshaller.Schedule{
				TrainIdentifiers: unmarshaller.TrainIdentifiers{
					RID:                "012345678901234",
					UID:                "A00001",
					ScheduledStartDate: "2025-08-23",
				},
				Headcode:         "2C04",
				RetailServiceID:  "NT123456",
				TOC:              "NT",
				Service:          railreader.ServiceBus,
				Category:         railreader.CategoryBusReplacement,
				PassengerService: true,
				Active:           true,
				Deleted:          true,
				Charter:          true,
				CancellationReason: &unmarshaller.DisruptionReason{
					TIPLOC:   "ABCD",
					Near:     true,
					ReasonID: 100,
				},
				DivertedVia: "EFGH",
				DiversionReason: &unmarshaller.DisruptionReason{
					TIPLOC:   "IJKL",
					Near:     true,
					ReasonID: 101,
				},
				Locations: []unmarshaller.ScheduleLocation{
					{
						Type: unmarshaller.LocationTypeOrigin,
						Origin: &unmarshaller.OriginLocation{
							LocationBase: unmarshaller.LocationBase{
								TIPLOC:              "MNOP",
								Activities:          "TBT ",
								PlannedActivities:   "TB",
								Cancelled:           true,
								FormationID:         "012345678901234-001",
								AffectedByDiversion: true,
								CancellationReason: &unmarshaller.DisruptionReason{
									TIPLOC:   "UVWX",
									Near:     true,
									ReasonID: 102,
								},
							},
							PublicArrivalTime:    "00:01",
							PublicDepartureTime:  "00:02",
							WorkingArrivalTime:   "00:03",
							WorkingDepartureTime: "00:04",
							FalseDestination:     "QRST",
						},
					},
					{
						Type: unmarshaller.LocationTypeOperationalOrigin,
						OperationalOrigin: &unmarshaller.OperationalOriginLocation{
							WorkingArrivalTime:   "00:05",
							WorkingDepartureTime: "00:06",
						},
					},
					{
						Type: unmarshaller.LocationTypeIntermediate,
						Intermediate: &unmarshaller.IntermediateLocation{
							PublicArrivalTime:    "00:07",
							PublicDepartureTime:  "00:08",
							WorkingArrivalTime:   "00:09",
							WorkingDepartureTime: "00:10",
							RoutingDelay:         2,
							FalseDestination:     "YZAB",
						},
					},
					{
						Type: unmarshaller.LocationTypeOperationalIntermediate,
						OperationalIntermediate: &unmarshaller.OperationalIntermediateLocation{
							WorkingArrivalTime:   "00:11",
							WorkingDepartureTime: "00:12",
							RoutingDelay:         3,
						},
					},
					{
						Type: unmarshaller.LocationTypeIntermediatePassing,
						IntermediatePassing: &unmarshaller.IntermediatePassingLocation{
							WorkingPassingTime: "00:13",
							RoutingDelay:       4,
						},
					},
					{
						Type: unmarshaller.LocationTypeDestination,
						Destination: &unmarshaller.DestinationLocation{
							PublicArrivalTime:    "00:14",
							PublicDepartureTime:  "00:15",
							WorkingArrivalTime:   "00:16",
							WorkingDepartureTime: "00:17",
							RoutingDelay:         5,
						},
					},
					{
						Type: unmarshaller.LocationTypeOperationalDestination,
						OperationalDestination: &unmarshaller.OperationalDestinationLocation{
							WorkingArrivalTime:   "00:18",
							WorkingDepartureTime: "00:19",
							RoutingDelay:         6,
						},
					},
				},
			},
			MessageID: "message2",
			ExpectedDatabaseSchedule: database.Schedule{
				ScheduleID:                     "012345678901234",
				MessageID:                      "message2",
				UID:                            "A00001",
				ScheduledStartDate:             time.Date(2025, 8, 23, 0, 0, 0, 0, location),
				Headcode:                       "2C04",
				RetailServiceID:                pointerTo("NT123456"),
				TrainOperatingCompanyID:        "NT",
				Service:                        string(railreader.ServiceBus),
				Category:                       string(railreader.CategoryBusReplacement),
				Active:                         true,
				Deleted:                        true,
				Charter:                        true,
				CancellationReasonID:           pointerTo(100),
				CancellationReasonLocationID:   pointerTo("ABCD"),
				CancellationReasonNearLocation: pointerTo(true),
				LateReasonID:                   pointerTo(101),
				LateReasonLocationID:           pointerTo("IJKL"),
				LateReasonNearLocation:         pointerTo(true),
				Locations: []database.ScheduleLocation{
					{
						Sequence:                       0,
						Type:                           string(unmarshaller.LocationTypeOrigin),
						LocationID:                     "MNOP",
						Activities:                     pointerTo("TBT "),
						PlannedActivities:              pointerTo("TB"),
						Cancelled:                      true,
						FormationID:                    pointerTo("012345678901234-001"),
						AffectedByDiversion:            true,
						CancellationReasonID:           pointerTo(102),
						CancellationReasonLocationID:   pointerTo("UVWX"),
						CancellationReasonNearLocation: pointerTo(true),
						PublicArrivalTime:              pointerTo(time.Date(2025, 8, 23, 0, 1, 0, 0, location)),
						PublicDepartureTime:            pointerTo(time.Date(2025, 8, 23, 0, 2, 0, 0, location)),
						WorkingArrivalTime:             pointerTo(time.Date(2025, 8, 23, 0, 3, 0, 0, location)),
						WorkingDepartureTime:           pointerTo(time.Date(2025, 8, 23, 0, 4, 0, 0, location)),
						FalseDestinationLocationID:     pointerTo("QRST"),
					},
					{
						Sequence:             1,
						Type:                 string(unmarshaller.LocationTypeOperationalOrigin),
						WorkingArrivalTime:   pointerTo(time.Date(2025, 8, 23, 0, 5, 0, 0, location)),
						WorkingDepartureTime: pointerTo(time.Date(2025, 8, 23, 0, 6, 0, 0, location)),
					},
					{
						Sequence:                   2,
						Type:                       string(unmarshaller.LocationTypeIntermediate),
						PublicArrivalTime:          pointerTo(time.Date(2025, 8, 23, 0, 7, 0, 0, location)),
						PublicDepartureTime:        pointerTo(time.Date(2025, 8, 23, 0, 8, 0, 0, location)),
						WorkingArrivalTime:         pointerTo(time.Date(2025, 8, 23, 0, 9, 0, 0, location)),
						WorkingDepartureTime:       pointerTo(time.Date(2025, 8, 23, 0, 10, 0, 0, location)),
						RoutingDelay:               pointerTo(2 * time.Minute),
						FalseDestinationLocationID: pointerTo("YZAB"),
					},
					{
						Sequence:             3,
						Type:                 string(unmarshaller.LocationTypeOperationalIntermediate),
						WorkingArrivalTime:   pointerTo(time.Date(2025, 8, 23, 0, 11, 0, 0, location)),
						WorkingDepartureTime: pointerTo(time.Date(2025, 8, 23, 0, 12, 0, 0, location)),
						RoutingDelay:         pointerTo(3 * time.Minute),
					},
					{
						Sequence:           4,
						Type:               string(unmarshaller.LocationTypeIntermediatePassing),
						WorkingPassingTime: pointerTo(time.Date(2025, 8, 23, 0, 13, 0, 0, location)),
						RoutingDelay:       pointerTo(4 * time.Minute),
					},
					{
						Sequence:             5,
						Type:                 string(unmarshaller.LocationTypeDestination),
						PublicArrivalTime:    pointerTo(time.Date(2025, 8, 23, 0, 14, 0, 0, location)),
						PublicDepartureTime:  pointerTo(time.Date(2025, 8, 23, 0, 15, 0, 0, location)),
						WorkingArrivalTime:   pointerTo(time.Date(2025, 8, 23, 0, 16, 0, 0, location)),
						WorkingDepartureTime: pointerTo(time.Date(2025, 8, 23, 0, 17, 0, 0, location)),
						RoutingDelay:         pointerTo(5 * time.Minute),
					},
					{
						Sequence:             6,
						Type:                 string(unmarshaller.LocationTypeOperationalDestination),
						WorkingArrivalTime:   pointerTo(time.Date(2025, 8, 23, 0, 18, 0, 0, location)),
						WorkingDepartureTime: pointerTo(time.Date(2025, 8, 23, 0, 19, 0, 0, location)),
						RoutingDelay:         pointerTo(6 * time.Minute),
					},
				},
			},
		},
		// Test that FormationIDs ripple between locations correctly
		//{},
	}
	for _, tc := range testCases {
		if err := interpretSchedule(log, tc.MessageID, sr, tc.Schedule); err != nil {
			t.Error(err)
			continue
		}
		actual, ok := sr.Schedules["012345678901234"]
		if !ok {
			t.Error("Schedule not added to repository")
			continue
		}
		if !cmp.Equal(actual, tc.ExpectedDatabaseSchedule) {
			t.Error(cmp.Diff(tc.ExpectedDatabaseSchedule, actual))
		}
	}
}
