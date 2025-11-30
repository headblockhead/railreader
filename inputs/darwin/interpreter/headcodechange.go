package interpreter

import (
	"github.com/google/uuid"
	"github.com/headblockhead/railreader/inputs/darwin/unmarshaller"
	"github.com/jackc/pgx/v5"
)

func (u UnitOfWork) interpretHeadcodeChange(headcodeChange unmarshaller.HeadcodeChange) error {
	record, err := u.headcodeChangeToRecord(headcodeChange)
	if err != nil {
		return err
	}
	err = u.insertHeadcodeChangeRecord(record)
	if err != nil {
		return err
	}
	return nil
}

type headcodeChangeRecord struct {
	ID uuid.UUID

	MessageID string

	OldHeadcode string
	NewHeadcode string

	TrainDescriber      string
	TrainDescriberBerth string
}

func (u UnitOfWork) headcodeChangeToRecord(headcodeChange unmarshaller.HeadcodeChange) (headcodeChangeRecord, error) {
	var record headcodeChangeRecord
	record.ID = uuid.New()
	record.MessageID = *u.messageID
	record.OldHeadcode = headcodeChange.OldHeadcode
	record.NewHeadcode = headcodeChange.NewHeadcode
	record.TrainDescriber = headcodeChange.TDLocation.Describer
	record.TrainDescriberBerth = headcodeChange.TDLocation.Berth
	return record, nil
}

func (u UnitOfWork) insertHeadcodeChangeRecord(record headcodeChangeRecord) error {
	_, err := u.tx.Exec(u.ctx, `
		INSERT INTO darwin.headcode_changes (
			id
			,message_id
			,old_headcode
			,new_headcode
			,train_describer
			,train_describer_berth
		) VALUES (
			@id
			,@message_id
			,@old_headcode
			,@new_headcode
			,@train_describer
			,@train_describer_berth
		)
	`, pgx.StrictNamedArgs{
		"id":                    record.ID,
		"message_id":            record.MessageID,
		"old_headcode":          record.OldHeadcode,
		"new_headcode":          record.NewHeadcode,
		"train_describer":       record.TrainDescriber,
		"train_describer_berth": record.TrainDescriberBerth,
	})
	return err
}
