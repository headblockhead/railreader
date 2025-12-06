package inserter

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/headblockhead/railreader/ingesters/darwin/unmarshaller"
	"github.com/jackc/pgx/v5"
)

func (u *UnitOfWork) InsertReference(reference unmarshaller.Reference, filename string) error {
	err := u.insertReferenceRecord(reference.ID, filename, time.Now().In(u.timezone))
	if err != nil {
		return fmt.Errorf("inserting reference record: %w", err)
	}

	locationRecords, err := u.locationsToRecords(reference.Locations, reference.ID)
	if err != nil {
		return err
	}
	err = u.copyNewLocationRecords(locationRecords)
	if err != nil {
		return err
	}

	tocRecords, err := u.tocsToRecords(reference.TrainOperatingCompanies, reference.ID)
	if err != nil {
		return err
	}
	err = u.copyNewTOCRecords(tocRecords)
	if err != nil {
		return err
	}

	lateReasonRecords, err := u.reasonsToRecords(reference.LateReasons, reference.ID)
	if err != nil {
		return err
	}
	err = u.copyNewLateReasonRecords(lateReasonRecords)
	if err != nil {
		return err
	}
	cancellationReasonRecords, err := u.reasonsToRecords(reference.CancellationReasons, reference.ID)
	if err != nil {
		return err
	}
	err = u.copyNewCancellationReasonRecords(cancellationReasonRecords)
	if err != nil {
		return err
	}

	viaConditionRecords, err := u.viaConditionsToRecords(reference.ViaConditions, reference.ID)
	if err != nil {
		return err
	}
	err = u.copyNewViaConditionRecords(viaConditionRecords)
	if err != nil {
		return err
	}

	cisRecords, err := u.cisToRecords(reference.CustomerInformationSystems, reference.ID)
	if err != nil {
		return err
	}
	err = u.copyNewCISRecords(cisRecords)
	if err != nil {
		return err
	}

	loadingCategoryRecords, err := u.loadingCategoriesToRecords(reference.LoadingCategories, reference.ID)
	if err != nil {
		return err
	}
	err = u.copyNewLoadingCategoryRecords(loadingCategoryRecords)
	if err != nil {
		return err
	}
	return nil
}

func (u *UnitOfWork) insertReferenceRecord(id string, filename string, importedAt time.Time) error {
	_, err := u.tx.Exec(u.ctx, `
			INSERT INTO darwin.reference_files (id, filename, imported_at, message_id)
			VALUES (@id, @filename, @imported_at, @message_id);
	`, pgx.StrictNamedArgs{
		"id":          id,
		"filename":    filename,
		"imported_at": importedAt,
		"message_id":  u.messageID,
	})
	if err != nil {
		return err
	}
	return nil
}

func (u *UnitOfWork) ReferenceFileAlreadyImported(filename string) (bool, error) {
	err := u.tx.QueryRow(u.ctx, `SELECT 1 FROM darwin.reference_files WHERE filename = @filename;`, pgx.StrictNamedArgs{
		"filename": filename,
	}).Scan(new(int))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

type locationRecord struct {
	ID         uuid.UUID
	LocationID string
	CRSid      *string
	TOCid      *string
	Name       *string

	ReferenceID string
}

func (u *UnitOfWork) locationsToRecords(locations []unmarshaller.LocationReference, referenceID string) ([]locationRecord, error) {
	records := make([]locationRecord, 0, len(locations))
	for _, location := range locations {
		var record locationRecord
		record.ID = uuid.New()
		record.LocationID = location.Location
		record.CRSid = location.CRS
		record.TOCid = location.TOC
		if location.Name != location.Location {
			record.Name = &location.Name
		}
		record.ReferenceID = referenceID
		records = append(records, record)
	}
	return records, nil
}

func (u *UnitOfWork) copyNewLocationRecords(locations []locationRecord) error {
	_, err := u.tx.CopyFrom(u.ctx, pgx.Identifier{"darwin", "locations"}, []string{"id", "location_id", "crs_id", "toc_id", "name", "reference_id"}, pgx.CopyFromSlice(len(locations), func(i int) ([]any, error) {
		location := locations[i]
		return []any{location.ID, location.LocationID, location.CRSid, location.TOCid, location.Name, location.ReferenceID}, nil
	}))
	return err
}

type tocRecord struct {
	ID    uuid.UUID
	TOCid string
	Name  string
	URL   *string

	ReferenceID string
}

func (u *UnitOfWork) tocsToRecords(tocs []unmarshaller.TrainOperatingCompanyReference, referenceID string) ([]tocRecord, error) {
	records := make([]tocRecord, 0, len(tocs))
	for _, toc := range tocs {
		var record tocRecord
		record.ID = uuid.New()
		record.TOCid = toc.ID
		record.Name = toc.Name
		if toc.URL != nil {
			record.URL = toc.URL
		}
		record.ReferenceID = referenceID
		records = append(records, record)
	}
	return records, nil
}

func (u *UnitOfWork) copyNewTOCRecords(tocs []tocRecord) error {
	_, err := u.tx.CopyFrom(u.ctx, pgx.Identifier{"darwin", "tocs"}, []string{"id", "toc_id", "name", "url", "reference_id"}, pgx.CopyFromSlice(len(tocs), func(i int) ([]any, error) {
		toc := tocs[i]
		return []any{toc.ID, toc.TOCid, toc.Name, toc.URL, toc.ReferenceID}, nil
	}))
	return err
}

type reasonRecord struct {
	ID uuid.UUID

	ReasonID    int
	Description string

	ReferenceID string
}

func (u *UnitOfWork) reasonsToRecords(reasons []unmarshaller.ReasonDescription, referenceID string) ([]reasonRecord, error) {
	records := make([]reasonRecord, 0, len(reasons))
	for _, reason := range reasons {
		var record reasonRecord
		record.ID = uuid.New()
		record.ReasonID = reason.ReasonID
		record.Description = reason.Description
		record.ReferenceID = referenceID
		records = append(records, record)
	}
	return records, nil
}

func (u *UnitOfWork) copyNewLateReasonRecords(reasons []reasonRecord) error {
	_, err := u.tx.CopyFrom(u.ctx, pgx.Identifier{"darwin", "late_reasons"}, []string{"id", "reason_id", "description", "reference_id"}, pgx.CopyFromSlice(len(reasons), func(i int) ([]any, error) {
		reason := reasons[i]
		return []any{reason.ID, reason.ReasonID, reason.Description, reason.ReferenceID}, nil
	}))
	return err
}

func (u *UnitOfWork) copyNewCancellationReasonRecords(reasons []reasonRecord) error {
	_, err := u.tx.CopyFrom(u.ctx, pgx.Identifier{"darwin", "cancellation_reasons"}, []string{"id", "reason_id", "description", "reference_id"}, pgx.CopyFromSlice(len(reasons), func(i int) ([]any, error) {
		reason := reasons[i]
		return []any{reason.ID, reason.ReasonID, reason.Description, reason.ReferenceID}, nil
	}))
	return err
}

type viaConditionRecord struct {
	ID uuid.UUID

	Sequence int

	DisplayAtCRSid                string
	FirstRequiredLocationID       string
	SecondRequiredLocationID      *string
	DestinationRequiredLocationID string
	Text                          string

	ReferenceID string
}

func (u *UnitOfWork) viaConditionsToRecords(viaConditions []unmarshaller.ViaCondition, referenceID string) ([]viaConditionRecord, error) {
	records := make([]viaConditionRecord, 0, len(viaConditions))
	for i, viaCondition := range viaConditions {
		var record viaConditionRecord
		record.ID = uuid.New()
		record.Sequence = i
		record.DisplayAtCRSid = viaCondition.DisplayAt
		record.FirstRequiredLocationID = viaCondition.RequiredCallingLocation1
		if viaCondition.RequiredCallingLocation2 != nil {
			record.SecondRequiredLocationID = viaCondition.RequiredCallingLocation2
		}
		record.DestinationRequiredLocationID = viaCondition.RequiredDestination
		record.Text = viaCondition.Text
		record.ReferenceID = referenceID
		records = append(records, record)
	}
	return records, nil
}

func (u *UnitOfWork) copyNewViaConditionRecords(viaConditions []viaConditionRecord) error {
	_, err := u.tx.CopyFrom(u.ctx, pgx.Identifier{"darwin", "via_conditions"}, []string{"id", "sequence", "display_at_crs_id", "first_required_location_id", "second_required_location_id", "destination_required_location_id", "text", "reference_id"}, pgx.CopyFromSlice(len(viaConditions), func(i int) ([]any, error) {
		viaCondition := viaConditions[i]
		return []any{viaCondition.ID, viaCondition.Sequence, viaCondition.DisplayAtCRSid, viaCondition.FirstRequiredLocationID, viaCondition.SecondRequiredLocationID, viaCondition.DestinationRequiredLocationID, viaCondition.Text, viaCondition.ReferenceID}, nil
	}))
	return err
}

type cisRecord struct {
	ID uuid.UUID

	CISid string
	Name  string

	ReferenceID string
}

func (u *UnitOfWork) cisToRecords(cis []unmarshaller.CISReference, referenceID string) ([]cisRecord, error) {
	records := make([]cisRecord, 0, len(cis))
	for _, ci := range cis {
		var record cisRecord
		record.ID = uuid.New()
		record.CISid = ci.CIS
		record.Name = ci.Name
		record.ReferenceID = referenceID
		records = append(records, record)
	}
	return records, nil
}
func (u *UnitOfWork) copyNewCISRecords(cis []cisRecord) error {
	_, err := u.tx.CopyFrom(u.ctx, pgx.Identifier{"darwin", "cis"}, []string{"id", "cis_id", "name", "reference_id"}, pgx.CopyFromSlice(len(cis), func(i int) ([]any, error) {
		ci := cis[i]
		return []any{ci.ID, ci.CISid, ci.Name, ci.ReferenceID}, nil
	}))
	return err
}

type loadingCategoryRecord struct {
	ID uuid.UUID

	Code  string
	TOCid *string

	Name                string
	DescriptionTypical  string
	DescriptionExpected string
	Definition          string

	ReferenceID string
}

func (u *UnitOfWork) loadingCategoriesToRecords(categories []unmarshaller.LoadingCategoryReference, referenceID string) ([]loadingCategoryRecord, error) {
	records := make([]loadingCategoryRecord, 0, len(categories))
	for _, category := range categories {
		var record loadingCategoryRecord
		record.ID = uuid.New()
		record.Code = category.Code
		record.TOCid = category.TOC
		record.Name = category.Name
		record.DescriptionTypical = category.ExpectedDescription
		record.DescriptionExpected = category.TypicalDescription
		record.Definition = category.Definition
		// colour and image fields go unused here
		record.ReferenceID = referenceID
		records = append(records, record)
	}
	return records, nil
}

func (u *UnitOfWork) copyNewLoadingCategoryRecords(categories []loadingCategoryRecord) error {
	_, err := u.tx.CopyFrom(u.ctx, pgx.Identifier{"darwin", "loading_categories"}, []string{"id", "code", "toc_id", "name", "description_typical", "description_expected", "definition", "reference_id"}, pgx.CopyFromSlice(len(categories), func(i int) ([]any, error) {
		category := categories[i]
		return []any{category.ID, category.Code, category.TOCid, category.Name, category.DescriptionTypical, category.DescriptionExpected, category.Definition, category.ReferenceID}, nil
	}))
	return err
}
