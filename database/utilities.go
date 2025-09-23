package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func DeleteAllInTable(ctx context.Context, tx pgx.Tx, tableName string) error {
	_, err := tx.Exec(ctx, `DELETE FROM @table_name;`, pgx.StrictNamedArgs{
		"table_name": tableName,
	})
	if err != pgx.ErrNoRows && err != nil {
		return fmt.Errorf("failed to delete existing %s: %w", tableName, err)
	}
	return nil
}
