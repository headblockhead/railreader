package interpreter

import (
	"fmt"

	"github.com/headblockhead/railreader/darwin/repository"
	"github.com/headblockhead/railreader/darwin/unmarshaller"
)

func (u UnitOfWork) InterpretReference(reference unmarshaller.Reference) error {
	u.log.Debug("interpreting a Reference")
	var locations []repository.LocationRow
	for _, loc := range reference.Locations {
		locations = append(locations, repository.LocationRow{
			LocationID:                      string(loc.Location),
			ComputerisedReservationSystemID: loc.CRS,
			TrainOperatingCompanyID:         loc.TOC,
			Name:                            loc.Name,
		})
	}
	if err := u.locationRepository.InsertMany(locations); err != nil {
		return fmt.Errorf("failed to insert locations: %w", err)
	}
	// TODO: add other reference data
	return nil
}
