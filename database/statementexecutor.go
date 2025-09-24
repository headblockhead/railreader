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

func InsertIntoTable[T any](ctx context.Context, tx pgx.Tx, tableName string, row T) error {
	statement, args := BuildInsert(tableName, row)
	if _, err := tx.Exec(ctx, statement, args); err != nil {
		return err
	}
	return nil
}

func InsertManyIntoTable[T any](ctx context.Context, tx pgx.Tx, tableName string, rows []T) error {
	if len(rows) == 0 {
		return nil
	}
	//var rowValues [][]any
	for _, row := range rows {
		if err := InsertIntoTable(ctx, tx, tableName, row); err != nil {
			return err
		}
		//rowValues = append(rowValues, values(reflect.ValueOf(row)))
	}
	/* _, err := tx.CopyFrom(ctx, pgx.Identifier{tableName}, columns(reflect.TypeFor[T]()), pgx.CopyFromRows(rowValues))*/
	/*if err != nil {*/
	/*return fmt.Errorf("failed to copy into %s: %w", tableName, err)*/
	/*}*/
	return nil
}
