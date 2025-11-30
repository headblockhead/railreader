package interpreter

import (
	"github.com/headblockhead/railreader/ingesters/darwin/unmarshaller"
	"github.com/jackc/pgx/v5"
)

func (u *UnitOfWork) InterpretReference(reference *unmarshaller.Reference) error {
	locationRecords := make([]locationRecord, 0, len(reference.Locations))
	for _, loc := range reference.Locations {
		record, err := u.locationToRecord(loc)
		if err != nil {
			return err
		}
		locationRecords = append(locationRecords, record)
	}
	if err := u.copyNewLocationRecords(locationRecords); err != nil {
		return err
	}
	return nil
}

type locationRecord struct {
	ID    string
	CRSid *string
	TOCid *string
	Name  *string
}

func (u *UnitOfWork) locationToRecord(location unmarshaller.LocationReference) (locationRecord, error) {
	var record locationRecord
	record.ID = location.Location
	record.CRSid = location.CRS
	record.TOCid = location.TOC
	if location.Name != location.Location {
		record.Name = &location.Name
	}
	return record, nil
}

func (u *UnitOfWork) copyNewLocationRecords(locations []locationRecord) error {
	_, err := u.tx.Exec(u.ctx, `
		DELETE FROM darwin.locations
	`)
	if err != nil {
		return err
	}

	_, err = u.tx.CopyFrom(u.ctx, pgx.Identifier{"darwin", "locations"}, []string{"id", "crs_id", "toc_id", "name"}, pgx.CopyFromSlice(len(locations), func(i int) ([]any, error) {
		loc := locations[i]
		return []any{loc.ID, loc.CRSid, loc.TOCid, loc.Name}, nil
	}))
	return err
}
