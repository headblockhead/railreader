package interpreter

import (
	"errors"
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
	Schedules map[string]database.ScheduleRow
}

func (sr *testingScheduleRepository) Insert(schedule database.ScheduleRow) error {
	sr.Schedules[schedule.ScheduleID] = schedule
	return nil
}

func (sr *testingScheduleRepository) Select(scheduleID string) (schedule database.ScheduleRow, err error) {
	schedule, ok := sr.Schedules[scheduleID]
	if !ok {
		err = errors.New("not found")
		return
	}
	return
}

func TestInterpretSchedule(t *testing.T) {
	sr := &testingScheduleRepository{Schedules: make(map[string]database.ScheduleRow)}
	log := slog.New(slog.NewTextHandler(t.Output(), &slog.HandlerOptions{Level: slog.LevelDebug}))
	location, err := time.LoadLocation("Europe/London")
	if err != nil {
		t.Fatalf("Failed to load location: %v", err)
	}
	testCases := map[string]struct {
		Schedule  unmarshaller.Schedule
		MessageID string

		ExpectedScheduleRow database.ScheduleRow
	}{
		// This is an example of a minimal valid schedule.
		"non_required_fields_are_nil": {
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
			ExpectedScheduleRow: database.ScheduleRow{
				ScheduleID:              "012345678901234",
				MessageID:               "message1",
				UID:                     "A00001",
				ScheduledStartDate:      time.Date(2025, 8, 23, 0, 0, 0, 0, location),
				Headcode:                "2C04",
				TrainOperatingCompanyID: "NT",
				Service:                 string(railreader.ServicePassengerOrParcelTrain),
				Category:                string(railreader.CategoryPassenger),
				PassengerService:        true,
				Active:                  true,
				Locations: []database.ScheduleLocationRow{
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
		// This is an imaginary schedule object where all possible fields have been changed from their defaults.
		"all_fields_are_stored": {
			Schedule: unmarshaller.Schedule{
				TrainIdentifiers: unmarshaller.TrainIdentifiers{
					RID:                "012345678901234",
					UID:                "A00001",
					ScheduledStartDate: "2025-08-23",
				},
				Headcode:         "2C04",
				RetailServiceID:  pointerTo("NT123456"),
				TOC:              "NT",
				Service:          railreader.ServiceBus,
				Category:         railreader.CategoryBusReplacement,
				PassengerService: true,
				Active:           true,
				Deleted:          true,
				Charter:          true,
				CancellationReason: &unmarshaller.DisruptionReason{
					TIPLOC:   pointerTo(railreader.TimingPointLocationCode("ABCD")),
					Near:     true,
					ReasonID: 100,
				},
				DivertedVia: pointerTo(railreader.TimingPointLocationCode("EFGH")),
				DiversionReason: &unmarshaller.DisruptionReason{
					TIPLOC:   pointerTo(railreader.TimingPointLocationCode("IJKL")),
					Near:     true,
					ReasonID: 101,
				},
				Locations: []unmarshaller.ScheduleLocation{
					{
						Type: unmarshaller.LocationTypeOrigin,
						Origin: &unmarshaller.OriginLocation{
							LocationBase: unmarshaller.LocationBase{
								TIPLOC:              "MNOP",
								Activities:          pointerTo("TBT "),
								PlannedActivities:   pointerTo("TB"),
								Cancelled:           true,
								FormationID:         pointerTo("012345678901234-001"),
								AffectedByDiversion: true,
								CancellationReason: &unmarshaller.DisruptionReason{
									TIPLOC:   pointerTo(railreader.TimingPointLocationCode("UVWX")),
									Near:     true,
									ReasonID: 102,
								},
							},
							PublicArrivalTime:    pointerTo(unmarshaller.TrainTime("00:01")),
							PublicDepartureTime:  pointerTo(unmarshaller.TrainTime("00:02")),
							WorkingArrivalTime:   pointerTo(unmarshaller.TrainTime("00:03")),
							WorkingDepartureTime: unmarshaller.TrainTime("00:04"),
							FalseDestination:     pointerTo(railreader.TimingPointLocationCode("QRST")),
						},
					},
					{
						Type: unmarshaller.LocationTypeOperationalOrigin,
						OperationalOrigin: &unmarshaller.OperationalOriginLocation{
							WorkingArrivalTime:   pointerTo(unmarshaller.TrainTime("00:05")),
							WorkingDepartureTime: "00:06",
						},
					},
					{
						Type: unmarshaller.LocationTypeIntermediate,
						Intermediate: &unmarshaller.IntermediateLocation{
							PublicArrivalTime:    pointerTo(unmarshaller.TrainTime("00:07")),
							PublicDepartureTime:  pointerTo(unmarshaller.TrainTime("00:08")),
							WorkingArrivalTime:   "00:09",
							WorkingDepartureTime: "00:10",
							RoutingDelay:         pointerTo(2),
							FalseDestination:     pointerTo(railreader.TimingPointLocationCode("YZAB")),
						},
					},
					{
						Type: unmarshaller.LocationTypeOperationalIntermediate,
						OperationalIntermediate: &unmarshaller.OperationalIntermediateLocation{
							WorkingArrivalTime:   "00:11",
							WorkingDepartureTime: "00:12",
							RoutingDelay:         pointerTo(3),
						},
					},
					{
						Type: unmarshaller.LocationTypeIntermediatePassing,
						IntermediatePassing: &unmarshaller.IntermediatePassingLocation{
							WorkingPassingTime: "00:13",
							RoutingDelay:       pointerTo(4),
						},
					},
					{
						Type: unmarshaller.LocationTypeDestination,
						Destination: &unmarshaller.DestinationLocation{
							PublicArrivalTime:    pointerTo(unmarshaller.TrainTime("00:14")),
							PublicDepartureTime:  pointerTo(unmarshaller.TrainTime("00:15")),
							WorkingArrivalTime:   "00:16",
							WorkingDepartureTime: pointerTo(unmarshaller.TrainTime("00:17")),
							RoutingDelay:         pointerTo(5),
						},
					},
					{
						Type: unmarshaller.LocationTypeOperationalDestination,
						OperationalDestination: &unmarshaller.OperationalDestinationLocation{
							WorkingArrivalTime:   "00:18",
							WorkingDepartureTime: pointerTo(unmarshaller.TrainTime("00:19")),
							RoutingDelay:         pointerTo(6),
						},
					},
				},
			},
			MessageID: "message2",
			ExpectedScheduleRow: database.ScheduleRow{
				ScheduleID:                     "012345678901234",
				MessageID:                      "message2",
				UID:                            "A00001",
				ScheduledStartDate:             time.Date(2025, 8, 23, 0, 0, 0, 0, location),
				Headcode:                       "2C04",
				RetailServiceID:                pointerTo("NT123456"),
				TrainOperatingCompanyID:        "NT",
				Service:                        string(railreader.ServiceBus),
				Category:                       string(railreader.CategoryBusReplacement),
				PassengerService:               true,
				Active:                         true,
				Deleted:                        true,
				Charter:                        true,
				CancellationReasonID:           pointerTo(100),
				CancellationReasonLocationID:   pointerTo("ABCD"),
				CancellationReasonNearLocation: pointerTo(true),
				LateReasonID:                   pointerTo(101),
				LateReasonLocationID:           pointerTo("IJKL"),
				LateReasonNearLocation:         pointerTo(true),
				Locations: []database.ScheduleLocationRow{
					{
						Sequence:                       0,
						Type:                           string(unmarshaller.LocationTypeOrigin),
						LocationID:                     "MNOP",
						Activities:                     pointerTo([]string{"TB", "T "}),
						PlannedActivities:              pointerTo([]string{"TB"}),
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
		"formationids_ripple": {
			Schedule: unmarshaller.Schedule{
				TrainIdentifiers: unmarshaller.TrainIdentifiers{
					ScheduledStartDate: "2025-08-23",
				},
				Locations: []unmarshaller.ScheduleLocation{
					// set a formationid
					{
						Type: unmarshaller.LocationTypeIntermediate,
						Intermediate: &unmarshaller.IntermediateLocation{
							LocationBase: unmarshaller.LocationBase{
								FormationID: pointerTo("012345678901234-001"),
							},
							WorkingArrivalTime:   "00:03",
							WorkingDepartureTime: "00:04",
						},
					},
					// test that a formationid will ripple to the next location
					{
						Type: unmarshaller.LocationTypeIntermediate,
						Intermediate: &unmarshaller.IntermediateLocation{
							LocationBase: unmarshaller.LocationBase{
								// should have formationid 012345678901234-001
							},
							WorkingArrivalTime:   "00:05",
							WorkingDepartureTime: "00:06",
						},
					},
					// test that a formationid will be nil if a cancelled location is given
					{
						Type: unmarshaller.LocationTypeIntermediate,
						Intermediate: &unmarshaller.IntermediateLocation{
							LocationBase: unmarshaller.LocationBase{
								Cancelled: true,
								// should have formationid = nil
							},
							WorkingArrivalTime:   "00:07",
							WorkingDepartureTime: "00:08",
						},
					},
					// set a formationid
					{
						Type: unmarshaller.LocationTypeIntermediate,
						Intermediate: &unmarshaller.IntermediateLocation{
							LocationBase: unmarshaller.LocationBase{
								Cancelled:   true,
								FormationID: pointerTo("012345678901234-002"),
							},
							WorkingArrivalTime:   "00:09",
							WorkingDepartureTime: "00:10",
						},
					},
					// test that cancelled locations do not ripple
					{
						Type: unmarshaller.LocationTypeIntermediate,
						Intermediate: &unmarshaller.IntermediateLocation{
							LocationBase: unmarshaller.LocationBase{
								// should have formationid = nil
							},
							WorkingArrivalTime:   "00:11",
							WorkingDepartureTime: "00:12",
						},
					},
					// test that a blank FID will result in nil
					{
						Type: unmarshaller.LocationTypeIntermediate,
						Intermediate: &unmarshaller.IntermediateLocation{
							LocationBase: unmarshaller.LocationBase{
								FormationID: pointerTo(""),
								// should have formationid = nil
							},
							WorkingArrivalTime:   "00:13",
							WorkingDepartureTime: "00:14",
						},
					},
					// test that this will also ripple onwards
					{
						Type: unmarshaller.LocationTypeIntermediate,
						Intermediate: &unmarshaller.IntermediateLocation{
							LocationBase: unmarshaller.LocationBase{
								// should have formationid = nil
							},
							WorkingArrivalTime:   "00:15",
							WorkingDepartureTime: "00:16",
						},
					},
				},
			},
			MessageID: "message3",
			ExpectedScheduleRow: database.ScheduleRow{
				MessageID:          "message3",
				ScheduledStartDate: time.Date(2025, 8, 23, 0, 0, 0, 0, location),
				Locations: []database.ScheduleLocationRow{
					{
						Sequence:             0,
						Type:                 string(unmarshaller.LocationTypeIntermediate),
						WorkingArrivalTime:   pointerTo(time.Date(2025, 8, 23, 0, 3, 0, 0, location)),
						WorkingDepartureTime: pointerTo(time.Date(2025, 8, 23, 0, 4, 0, 0, location)),
						FormationID:          pointerTo("012345678901234-001"),
					},
					{
						Sequence:             1,
						Type:                 string(unmarshaller.LocationTypeIntermediate),
						WorkingArrivalTime:   pointerTo(time.Date(2025, 8, 23, 0, 5, 0, 0, location)),
						WorkingDepartureTime: pointerTo(time.Date(2025, 8, 23, 0, 6, 0, 0, location)),
						FormationID:          pointerTo("012345678901234-001"),
					},
					{
						Sequence:             2,
						Type:                 string(unmarshaller.LocationTypeIntermediate),
						WorkingArrivalTime:   pointerTo(time.Date(2025, 8, 23, 0, 7, 0, 0, location)),
						WorkingDepartureTime: pointerTo(time.Date(2025, 8, 23, 0, 8, 0, 0, location)),
						Cancelled:            true,
					},
					{
						Sequence:             3,
						Type:                 string(unmarshaller.LocationTypeIntermediate),
						WorkingArrivalTime:   pointerTo(time.Date(2025, 8, 23, 0, 9, 0, 0, location)),
						WorkingDepartureTime: pointerTo(time.Date(2025, 8, 23, 0, 10, 0, 0, location)),
						Cancelled:            true,
						FormationID:          pointerTo("012345678901234-002"),
					},
					{
						Sequence:             4,
						Type:                 string(unmarshaller.LocationTypeIntermediate),
						WorkingArrivalTime:   pointerTo(time.Date(2025, 8, 23, 0, 11, 0, 0, location)),
						WorkingDepartureTime: pointerTo(time.Date(2025, 8, 23, 0, 12, 0, 0, location)),
					},
					{
						Sequence:             5,
						Type:                 string(unmarshaller.LocationTypeIntermediate),
						WorkingArrivalTime:   pointerTo(time.Date(2025, 8, 23, 0, 13, 0, 0, location)),
						WorkingDepartureTime: pointerTo(time.Date(2025, 8, 23, 0, 14, 0, 0, location)),
					},
					{
						Sequence:             6,
						Type:                 string(unmarshaller.LocationTypeIntermediate),
						WorkingArrivalTime:   pointerTo(time.Date(2025, 8, 23, 0, 15, 0, 0, location)),
						WorkingDepartureTime: pointerTo(time.Date(2025, 8, 23, 0, 16, 0, 0, location)),
					},
				},
			},
		},
		"empty_activities_produces_activitynone": {
			Schedule: unmarshaller.Schedule{
				TrainIdentifiers: unmarshaller.TrainIdentifiers{
					ScheduledStartDate: "2025-08-23",
				},
				Locations: []unmarshaller.ScheduleLocation{
					{
						Type: unmarshaller.LocationTypeIntermediate,
						Intermediate: &unmarshaller.IntermediateLocation{
							LocationBase: unmarshaller.LocationBase{
								Activities: pointerTo(""),
							},
							WorkingArrivalTime:   "00:03",
							WorkingDepartureTime: "00:04",
						},
					},
				},
			},
			MessageID: "message4",
			ExpectedScheduleRow: database.ScheduleRow{
				MessageID:          "message4",
				ScheduledStartDate: time.Date(2025, 8, 23, 0, 0, 0, 0, location),
				Locations: []database.ScheduleLocationRow{
					{
						Sequence:             0,
						Type:                 string(unmarshaller.LocationTypeIntermediate),
						Activities:           pointerTo([]string{"  "}),
						WorkingArrivalTime:   pointerTo(time.Date(2025, 8, 23, 0, 3, 0, 0, location)),
						WorkingDepartureTime: pointerTo(time.Date(2025, 8, 23, 0, 4, 0, 0, location)),
					},
				},
			},
		},
		"dates_are_correct_when_times_cross_midnight": {
			Schedule: unmarshaller.Schedule{
				TrainIdentifiers: unmarshaller.TrainIdentifiers{
					ScheduledStartDate: "2025-08-23",
				},
				Locations: []unmarshaller.ScheduleLocation{
					// Test crossing midnight forward inside a location, as the first location
					{
						Type: unmarshaller.LocationTypeIntermediate,
						Intermediate: &unmarshaller.IntermediateLocation{
							WorkingArrivalTime:   "23:59",
							WorkingDepartureTime: "00:01",
						},
					},
					// Test crossing midnight backward
					{
						Type: unmarshaller.LocationTypeIntermediate,
						Intermediate: &unmarshaller.IntermediateLocation{
							WorkingArrivalTime:   "23:55",
							WorkingDepartureTime: "23:56",
						},
					},
					// Test crossing midnight forward
					{
						Type: unmarshaller.LocationTypeIntermediate,
						Intermediate: &unmarshaller.IntermediateLocation{
							WorkingArrivalTime:   "00:05",
							WorkingDepartureTime: "00:06",
						},
					},
					// Test backwards time
					{
						Type: unmarshaller.LocationTypeIntermediate,
						Intermediate: &unmarshaller.IntermediateLocation{
							WorkingArrivalTime:   "23:56",
							WorkingDepartureTime: "23:57",
						},
					},
					// Test forwards time
					{
						Type: unmarshaller.LocationTypeIntermediate,
						Intermediate: &unmarshaller.IntermediateLocation{
							WorkingArrivalTime:   "23:58",
							WorkingDepartureTime: "23:59",
						},
					},
					// Test crossing midnight forward in a location, not the first location
					{
						Type: unmarshaller.LocationTypeIntermediate,
						Intermediate: &unmarshaller.IntermediateLocation{
							WorkingArrivalTime:   "23:59",
							WorkingDepartureTime: "00:01",
						},
					},
				},
			},
			MessageID: "message5",
			ExpectedScheduleRow: database.ScheduleRow{
				MessageID:          "message5",
				ScheduledStartDate: time.Date(2025, 8, 23, 0, 0, 0, 0, location),
				Locations: []database.ScheduleLocationRow{
					{
						Sequence:             0,
						Type:                 string(unmarshaller.LocationTypeIntermediate),
						WorkingArrivalTime:   pointerTo(time.Date(2025, 8, 23, 23, 59, 0, 0, location)),
						WorkingDepartureTime: pointerTo(time.Date(2025, 8, 24, 0, 1, 0, 0, location)),
					},
					{
						Sequence:             1,
						Type:                 string(unmarshaller.LocationTypeIntermediate),
						WorkingArrivalTime:   pointerTo(time.Date(2025, 8, 23, 23, 55, 0, 0, location)),
						WorkingDepartureTime: pointerTo(time.Date(2025, 8, 23, 23, 56, 0, 0, location)),
					},
					{
						Sequence:             2,
						Type:                 string(unmarshaller.LocationTypeIntermediate),
						WorkingArrivalTime:   pointerTo(time.Date(2025, 8, 24, 0, 5, 0, 0, location)),
						WorkingDepartureTime: pointerTo(time.Date(2025, 8, 24, 0, 6, 0, 0, location)),
					},
					{
						Sequence:             3,
						Type:                 string(unmarshaller.LocationTypeIntermediate),
						WorkingArrivalTime:   pointerTo(time.Date(2025, 8, 23, 23, 56, 0, 0, location)),
						WorkingDepartureTime: pointerTo(time.Date(2025, 8, 23, 23, 57, 0, 0, location)),
					},
					{
						Sequence:             4,
						Type:                 string(unmarshaller.LocationTypeIntermediate),
						WorkingArrivalTime:   pointerTo(time.Date(2025, 8, 23, 23, 58, 0, 0, location)),
						WorkingDepartureTime: pointerTo(time.Date(2025, 8, 23, 23, 59, 0, 0, location)),
					},
					{
						Sequence:             5,
						Type:                 string(unmarshaller.LocationTypeIntermediate),
						WorkingArrivalTime:   pointerTo(time.Date(2025, 8, 23, 23, 59, 0, 0, location)),
						WorkingDepartureTime: pointerTo(time.Date(2025, 8, 24, 0, 1, 0, 0, location)),
					},
				},
			},
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			if err := interpretSchedule(log, tc.MessageID, sr, tc.Schedule); err != nil {
				t.Fatal(err)
			}
			actual, ok := sr.Schedules[tc.Schedule.RID]
			if !ok {
				t.Error("Schedule not added to repository")
			}
			if !cmp.Equal(actual, tc.ExpectedScheduleRow) {
				t.Error(cmp.Diff(actual, tc.ExpectedScheduleRow))
			}
		})
	}
}
