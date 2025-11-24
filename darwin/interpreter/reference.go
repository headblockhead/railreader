package interpreter

import (
	"github.com/headblockhead/railreader/darwin/unmarshaller"
)

func (u *UnitOfWork) InterpretReference(reference unmarshaller.Reference) error {

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
