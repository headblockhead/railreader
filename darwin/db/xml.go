package db

import (
	"fmt"

	"github.com/jackc/pgx/v5"
)

func (c *Connection) InsertXML(messageID string, xmlData string) error {
	tx, err := c.pgxConnection.Begin(c.context)
	if err != nil {
		return fmt.Errorf("failed to begin transaction while inserting XML data: %w", err)
	}
	defer tx.Rollback(c.context)

	if _, err := tx.Exec(c.context, `
		INSERT INTO message_ids_xml
			VALUES (
				@message_id,
				@xml_data
			)
	`, pgx.StrictNamedArgs{
		"message_id": messageID,
		"xml_data":   xmlData,
	}); err != nil {
		return fmt.Errorf("failed to insert XML data for message %s: %w", messageID, err)
	}

	if err := tx.Commit(c.context); err != nil {
		return err
	}

	return nil
}
