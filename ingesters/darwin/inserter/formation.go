package inserter

import (
	"github.com/google/uuid"
	"github.com/headblockhead/railreader/ingesters/darwin/unmarshaller"
	"github.com/jackc/pgx/v5"
)

func (u *UnitOfWork) insertFormation(formation unmarshaller.FormationsOfService) error {
	fRecords, cRecords, err := u.formationsToRecords(formation)
	if err != nil {
		return err
	}
	err = u.insertFormationRecords(fRecords)
	if err != nil {
		return err
	}
	err = u.insertFormationCoachRecords(cRecords)
	if err != nil {
		return err
	}
	return nil
}

type formationRecord struct {
	ID           uuid.UUID
	MessageID    string
	ScheduleID   string
	FormationID  string
	Source       *string
	SourceSystem *string
}

type formationCoachRecord struct {
	ID            uuid.UUID
	FormationUUID uuid.UUID

	FormationID  string
	Identifier   string
	Class        *string
	ToiletType   string
	ToiletStatus string
}

func (u *UnitOfWork) formationsToRecords(formations unmarshaller.FormationsOfService) ([]formationRecord, []formationCoachRecord, error) {
	var fRecords []formationRecord
	var cRecords []formationCoachRecord
	for _, f := range formations.Formations {
		var fRecord formationRecord
		fRecord.ID = uuid.New()
		fRecord.MessageID = *u.messageID
		fRecord.ScheduleID = formations.RID
		fRecord.FormationID = f.ID
		fRecord.Source = f.Source
		fRecord.SourceSystem = f.SourceSystem
		fRecords = append(fRecords, fRecord)
		for _, c := range f.Coaches {
			var cRecord formationCoachRecord
			cRecord.ID = uuid.New()
			cRecord.FormationUUID = fRecord.ID
			cRecord.FormationID = f.ID
			cRecord.Identifier = c.Identifier
			cRecord.Class = c.Class
			cRecord.ToiletType = c.Toilet.Type
			cRecord.ToiletStatus = c.Toilet.Status
			cRecords = append(cRecords, cRecord)
		}
	}
	return fRecords, cRecords, nil
}

func (u *UnitOfWork) insertFormationRecords(records []formationRecord) error {
	for _, record := range records {
		u.batch.Queue(`
			INSERT INTO darwin.formations (
				id
				,message_id
				,schedule_id
				,formation_id
				,source
				,source_system
			) VALUES (
				@id
				,@message_id
				,@schedule_id
				,@formation_id
				,@source
				,@source_system
			);
			`, pgx.StrictNamedArgs{
			"id":            record.ID,
			"message_id":    record.MessageID,
			"schedule_id":   record.ScheduleID,
			"formation_id":  record.FormationID,
			"source":        record.Source,
			"source_system": record.SourceSystem,
		})
	}
	return nil
}

func (u *UnitOfWork) insertFormationCoachRecords(records []formationCoachRecord) error {
	for _, record := range records {
		u.batch.Queue(`
			INSERT INTO darwin.formation_coaches (
				id
				,formation_uuid
				,formation_id
				,identifier
				,class
				,toilet_type
				,toilet_status
			) VALUES (
				@id
				,@formation_uuid
				,@formation_id
				,@identifier
				,@class
				,@toilet_type
				,@toilet_status
			);
			`, pgx.StrictNamedArgs{
			"id":             record.ID,
			"formation_uuid": record.FormationUUID,
			"formation_id":   record.FormationID,
			"identifier":     record.Identifier,
			"class":          record.Class,
			"toilet_type":    record.ToiletType,
			"toilet_status":  record.ToiletStatus,
		})
	}
	return nil
}
