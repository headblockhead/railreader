package interpreter

import (
	"fmt"

	"github.com/headblockhead/railreader/darwin/repository"
	"github.com/headblockhead/railreader/darwin/unmarshaller"
)

func (u UnitOfWork) GetLastReference() (*repository.ReferenceRow, error) {
	return u.referenceRepository.SelectLast()
}

func (u UnitOfWork) InterpretReference(reference unmarshaller.Reference, filename string) error {
	u.log.Debug("interpreting a Reference")

	if err := u.referenceRepository.Insert(repository.ReferenceRow{
		Filename: filename,
	}); err != nil {
		return fmt.Errorf("failed to insert reference record: %w", err)
	}

	var locations []repository.LocationRow
	for _, loc := range reference.Locations {
		row := repository.LocationRow{}
		row.LocationID = string(loc.Location)
		row.ComputerisedReservationSystemID = loc.CRS
		row.TrainOperatingCompanyID = loc.TOC
		row.Name = loc.Name
		locations = append(locations, row)
	}
	if err := u.locationRepository.DeleteAll(); err != nil {
		return fmt.Errorf("failed to delete existing locations: %w", err)
	}
	if err := u.locationRepository.InsertMany(locations); err != nil {
		return fmt.Errorf("failed to insert locations: %w", err)
	}

	var tocs []repository.TrainOperatingCompanyRow
	for _, toc := range reference.TrainOperatingCompanies {
		row := repository.TrainOperatingCompanyRow{}
		row.TrainOperatingCompanyID = toc.ID
		row.Name = toc.Name
		row.URL = toc.URL
		tocs = append(tocs, row)
	}
	if err := u.trainOperatingCompanyRepository.DeleteAll(); err != nil {
		return fmt.Errorf("failed to delete existing train operating companies: %w", err)
	}
	if err := u.trainOperatingCompanyRepository.InsertMany(tocs); err != nil {
		return fmt.Errorf("failed to insert train operating companies: %w", err)
	}

	var lateReasons []repository.LateReasonRow
	for _, reason := range reference.LateReasons {
		row := repository.LateReasonRow{}
		row.LateReasonID = reason.ReasonID
		row.Description = reason.Description
		lateReasons = append(lateReasons, row)
	}
	if err := u.lateReasonRepository.DeleteAll(); err != nil {
		return fmt.Errorf("failed to delete existing late reasons: %w", err)
	}
	if err := u.lateReasonRepository.InsertMany(lateReasons); err != nil {
		return fmt.Errorf("failed to insert late reasons: %w", err)
	}

	var cancellationReasons []repository.CancellationReasonRow
	for _, reason := range reference.CancellationReasons {
		row := repository.CancellationReasonRow{}
		row.CancellationReasonID = reason.ReasonID
		row.Description = reason.Description
		cancellationReasons = append(cancellationReasons, row)
	}
	if err := u.cancellationReasonRepository.DeleteAll(); err != nil {
		return fmt.Errorf("failed to delete existing cancellation reasons: %w", err)
	}
	if err := u.cancellationReasonRepository.InsertMany(cancellationReasons); err != nil {
		return fmt.Errorf("failed to insert cancellation reasons: %w", err)
	}

	var viaConditions []repository.ViaConditionRow
	for i, via := range reference.ViaConditions {
		row := repository.ViaConditionRow{}
		row.Sequence = i
		row.DisplayAtComputerisedReservationSystemID = via.DisplayAt
		row.FirstRequiredLocationID = string(via.RequiredCallingLocation1)
		if via.RequiredCallingLocation2 != nil {
			row.SecondRequiredLocationID = pointerTo(string(*via.RequiredCallingLocation2))
		}
		row.DestinationRequiredLocationID = string(via.RequiredDestination)
		row.Text = via.Text
		viaConditions = append(viaConditions, row)
	}
	if err := u.viaConditionRepository.DeleteAll(); err != nil {
		return fmt.Errorf("failed to delete existing via conditions: %w", err)
	}
	if err := u.viaConditionRepository.InsertMany(viaConditions); err != nil {
		return fmt.Errorf("failed to insert via conditions: %w", err)
	}

	var customerInformationSystems []repository.CustomerInformationSystemRow
	for _, cis := range reference.CustomerInformationSystems {
		row := repository.CustomerInformationSystemRow{}
		row.CustomerInformationSystemID = cis.CIS
		row.Name = cis.Name
		customerInformationSystems = append(customerInformationSystems, row)
	}
	if err := u.customerInformationSystemRepository.DeleteAll(); err != nil {
		return fmt.Errorf("failed to delete existing customer information systems: %w", err)
	}
	if err := u.customerInformationSystemRepository.InsertMany(customerInformationSystems); err != nil {
		return fmt.Errorf("failed to insert customer information systems: %w", err)
	}

	var loadingCategories []repository.LoadingCategoryRow
	for _, lc := range reference.LoadingCategories {
		row := repository.LoadingCategoryRow{}
		row.LoadingCategoryCode = lc.Code
		row.TrainOperatingCompanyID = lc.TOC
		row.Name = lc.Name
		row.DescriptionTypical = lc.TypicalDescription
		row.DescriptionExpected = lc.ExpectedDescription
		row.Definition = lc.Definition
		loadingCategories = append(loadingCategories, row)
	}
	if err := u.loadingCategoryRepository.DeleteAll(); err != nil {
		return fmt.Errorf("failed to delete existing loading categories: %w", err)
	}
	for _, lc := range loadingCategories {
		if err := u.loadingCategoryRepository.Insert(lc); err != nil {
			return fmt.Errorf("failed to insert loading category: %w", err)
		}
	}
	return nil
}
