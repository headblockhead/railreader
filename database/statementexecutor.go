package database

import (
	"context"
	"fmt"
	"reflect"

	"github.com/jackc/pgx/v5"
)

// row is only used for its type information, not its value
func SelectFromTable[T any](ctx context.Context, tx pgx.Tx, tableName string, row T, whereClause string, whereArgs pgx.StrictNamedArgs) ([]T, error) {
	statement, args := BuildSelect(tableName, row, whereClause, whereArgs)
	rows, err := tx.Query(ctx, statement, args)
	if err != nil {
		return nil, fmt.Errorf("failed to select from %s: %w", tableName, err)
	}
	results, err := pgx.CollectRows(rows, pgx.RowToStructByName[T])
	if err != nil {
		return nil, fmt.Errorf("failed to collect rows from %s: %w", tableName, err)
	}
	return results, nil
}

// row is only used for its type information, not its value
func SelectOneFromTable[T any](ctx context.Context, tx pgx.Tx, tableName string, row T, whereClause string, whereArgs pgx.StrictNamedArgs) (T, error) {
	statement, args := BuildOneSelect(tableName, row, whereClause, whereArgs)
	rows, err := tx.Query(ctx, statement, args)
	var zero T
	if err != nil {
		return zero, fmt.Errorf("failed to select one from %s: %w", tableName, err)
	}
	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[T])
	if err != nil {
		return zero, fmt.Errorf("failed to collect exactly one row from %s: %w", tableName, err)
	}
	return result, nil
}

func DeleteAllInTable(ctx context.Context, tx pgx.Tx, tableName string) error {
	_, err := tx.Exec(ctx, `TRUNCATE TABLE `+tableName+`;`)
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
	// Insert using COPY for performance, and ignoring existing rows

	_, err := tx.Exec(ctx, `
		CREATE TEMPORARY TABLE temp_`+tableName+`
		ON COMMIT DROP
		AS SELECT * FROM `+tableName+` WITH NO DATA;
	`)
	if err != nil {
		return fmt.Errorf("failed to create temp_%s: %w", tableName, err)
	}

	var rowValues [][]any
	for _, row := range rows {
		rowValues = append(rowValues, values(reflect.ValueOf(row)))
	}
	_, err = tx.CopyFrom(ctx, pgx.Identifier{"temp_" + tableName}, columns(reflect.TypeFor[T]()), pgx.CopyFromRows(rowValues))
	if err != nil {
		return fmt.Errorf("failed to copy into %s: %w", tableName, err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO `+tableName+`
		SELECT * FROM temp_`+tableName+`
		ON CONFLICT DO NOTHING;
	`)
	if err != nil {
		return fmt.Errorf("failed to insert from temp_%s into %s: %w", tableName, tableName, err)
	}

	_, err = tx.Exec(ctx, `DROP TABLE temp_`+tableName+`;`)
	if err != nil {
		return fmt.Errorf("failed to drop temp_%s: %w", tableName, err)
	}

	return nil
}
