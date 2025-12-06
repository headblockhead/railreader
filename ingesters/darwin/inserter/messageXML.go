package inserter

import "github.com/jackc/pgx/v5"

type MessageXMLRecord struct {
	ID            string
	KafkaOffset   int64
	PPortSequence string
	XML           string
}

func (u *UnitOfWork) InsertMessageXMLRecord(record MessageXMLRecord) error {
	_, err := u.tx.Exec(u.ctx, `
		INSERT INTO darwin.message_xml (
			id
			,kafka_offset
			,pport_sequence
			,xml
		) VALUES (
			@id
			,@kafka_offset
			,@pport_sequence
			,@xml
		);
	`, pgx.StrictNamedArgs{
		"id":             record.ID,
		"kafka_offset":   record.KafkaOffset,
		"pport_sequence": record.PPortSequence,
		"xml":            record.XML,
	})
	if err != nil {
		return err
	}
	return nil
}
