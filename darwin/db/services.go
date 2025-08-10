package db

import (
	"fmt"
	"log/slog"
	"time"
)

// ProcessService takes a Service struct and creates/updates/deletes records appropriately.
func (c *Connection) ProcessService(s Service) error {
	log := c.log.With(slog.String("service_id", s.ServiceID))

	log.Debug("processing service")
	tx, err := c.connection.Begin(c.context)
	if err != nil {
		return fmt.Errorf("failed to begin transaction while processing a service: %w", err)
	}

	for _, loc := range s.Locations {
		if err := c.processServiceLocation(s.ServiceID, &loc); err != nil {
			return fmt.Errorf("failed to process location %s for service %s: %w", loc.LocationID, s.ServiceID, err)
		}
	}

	if err := tx.Commit(c.context); err != nil {
		return fmt.Errorf("failed to commit transaction while processing a service: %w", err)
	}
	return nil
}

func (c *Connection) processServiceLocation(serviceID string, location *ServiceLocation) error {
	return nil
}

type Service struct {
	ServiceID string

	UID                     string
	ScheduledStartDate      time.Time
	Headcode                string
	RetailServiceID         string
	TrainOperatingCompanyID string
	Service                 string
	Category                string
	Active                  bool
	Deleted                 bool
	Charter                 bool

	CancellationReasonID           int
	CancellationReasonLocationID   string
	CancellationReasonNearLocation bool

	LateReasonID           int
	LateReasonLocationID   string
	LateReasonNearLocation bool

	DivertedViaLocationID string

	Locations []ServiceLocation
}

type ServiceLocation struct {
	// ServiceID  string
	LocationID string
	Sequence   int

	Activities          string
	PlannedActivities   string
	Cancelled           bool
	AffectedByDiversion bool

	Type                       string
	PublicArrivalTime          time.Time
	PublicDepartureTime        time.Time
	WorkingArrivalTime         time.Time
	WorkingPassingTime         time.Time
	WorkingDepartureTime       time.Time
	RoutingDelay               time.Duration
	FalseDestinationLocationID string

	CancellationReasonID           int
	CancellationReasonLocationID   string
	CancellationReasonNearLocation bool
}
