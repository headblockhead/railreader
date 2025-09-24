package database

import (
	"maps"
	"reflect"
	"strings"

	"github.com/jackc/pgx/v5"
)

func forEachColumn(t reflect.Type, fn func(column string)) {
	for i := 0; i < t.NumField(); i++ {
		tag := t.Field(i).Tag.Get("db")
		if tag == "" || tag == "-" {
			continue
		}
		fn(tag)
	}
}

func columns(t reflect.Type) (columns []string) {
	forEachColumn(t, func(column string) {
		columns = append(columns, column)
	})
	return
}

func forEachValue(v reflect.Value, fn func(value any)) {
	for i := 0; i < v.NumField(); i++ {
		fn(v.Field(i).Interface())
	}
}

func values(v reflect.Value) (values []any) {
	forEachValue(v, func(value any) {
		values = append(values, value)
	})
	return
}

func forEachColumnValue(t reflect.Type, v reflect.Value, fn func(column string, value any)) {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		dbTag := field.Tag.Get("db")
		if dbTag == "" || dbTag == "-" {
			continue
		}
		fn(dbTag, v.Field(i).Interface())
	}
}

func columnsAndArgs(t reflect.Type, v reflect.Value) (columns []string, args pgx.StrictNamedArgs) {
	args = pgx.StrictNamedArgs{}
	forEachColumnValue(t, v, func(column string, value any) {
		columns = append(columns, column)
		args[column] = value
	})
	return
}

func BuildSelect[T any](tableName string, row T, whereClause string, whereArgs pgx.StrictNamedArgs) (sql string, args pgx.StrictNamedArgs) {
	columns := columns(reflect.TypeOf(row))
	sql = "SELECT " + strings.Join(columns, ", ") + " FROM " + tableName
	if whereClause != "" {
		sql += " WHERE " + whereClause
	}
	sql += ";"
	return sql, whereArgs
}

func BuildInsert[T any](tableName string, row T) (sql string, args pgx.StrictNamedArgs) {
	columns, args := columnsAndArgs(reflect.TypeOf(row), reflect.ValueOf(row))
	sql = "INSERT INTO " + tableName + " (" + strings.Join(columns, ", ") + ") VALUES (" + strings.Join(columns, ", @") + ") ON CONFLICT DO NOTHING;"
	return sql, args
}

func BuildUpdate[T any](tableName string, row T, whereClause string, whereArgs pgx.StrictNamedArgs) (sql string, args pgx.StrictNamedArgs) {
	columns, args := columnsAndArgs(reflect.TypeOf(row), reflect.ValueOf(row))
	maps.Copy(args, whereArgs)

	setClauses := make([]string, len(columns))
	for i, column := range columns {
		setClauses[i] = column + " = @" + column
	}

	sql = "UPDATE " + tableName + " SET " + strings.Join(setClauses, ", ")
	if whereClause != "" {
		sql += " WHERE " + whereClause
	}
	sql += ";"
	return sql, args
}

func BuildDelete(tableName string, whereClause string, whereArgs pgx.StrictNamedArgs) (sql string, args pgx.StrictNamedArgs) {
	sql = "DELETE FROM " + tableName
	if whereClause != "" {
		sql += " WHERE " + whereClause
	}
	sql += ";"
	return sql, whereArgs
}
