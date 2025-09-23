package database

import (
	"maps"
	"reflect"
	"strings"

	"github.com/jackc/pgx/v5"
)

func forEachColumn(t reflect.Type, v reflect.Value, fn func(column string, value any)) {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		dbTag := field.Tag.Get("db")
		if dbTag == "" || dbTag == "-" {
			continue
		}
		fn(dbTag, v.Field(i).Interface())
	}
}

func BuildInsert(tableName string, row any) (string, pgx.StrictNamedArgs) {
	t := reflect.TypeOf(row)
	v := reflect.ValueOf(row)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
		v = v.Elem()
	}

	columns := []string{}
	namedArgs := pgx.StrictNamedArgs{}
	forEachColumn(t, v, func(column string, value any) {
		namedArgs[column] = value
		columns = append(columns, column)
	})

	sql := "INSERT INTO " + tableName + " (" + strings.Join(columns, ", ") + ") VALUES (" + strings.Join(columns, ", @") + ");"
	return sql, namedArgs
}

func BuildUpdate(tableName string, row any, whereClause string, whereArgs pgx.StrictNamedArgs) (string, pgx.StrictNamedArgs) {
	v := reflect.ValueOf(row)
	t := reflect.TypeOf(row)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
		v = v.Elem()
	}

	columns := []string{}
	namedArgs := pgx.StrictNamedArgs{}
	forEachColumn(t, v, func(column string, value any) {
		columns = append(columns, column+" = @"+column)
		namedArgs[column] = value
	})

	maps.Copy(namedArgs, whereArgs)
	setString := strings.Join(columns, ", ")
	sql := "UPDATE " + tableName + " SET " + setString
	if whereClause != "" {
		sql += " WHERE " + whereClause
	}
	sql += ";"

	return sql, namedArgs
}
