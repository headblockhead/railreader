package database

import (
	"maps"
	"reflect"
	"strings"

	"github.com/jackc/pgx/v5"
)

// TODO: read through and improve these 'build' functions

func BuildInsert(tableName string, row any) (string, pgx.StrictNamedArgs) {
	v := reflect.ValueOf(row)
	t := reflect.TypeOf(row)

	if t.Kind() == reflect.Pointer {
		t = t.Elem()
		v = v.Elem()
	}

	var columns []string
	var placeholders []string
	namedArgs := pgx.StrictNamedArgs{}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		dbTag := field.Tag.Get("db")
		if dbTag == "" || dbTag == "-" {
			continue
		}
		columns = append(columns, dbTag)
		placeholders = append(placeholders, "@"+dbTag)
		namedArgs[dbTag] = v.Field(i).Interface()
	}

	columnsStr := strings.Join(columns, ", ")
	placeholdersStr := strings.Join(placeholders, ", ")

	sql := "INSERT INTO " + tableName + " (" + columnsStr + ") VALUES (" + placeholdersStr + ");"
	return sql, namedArgs
}

func BuildUpdate(tableName string, row any, whereClause string, whereArgs pgx.StrictNamedArgs) (string, pgx.StrictNamedArgs) {
	v := reflect.ValueOf(row)
	t := reflect.TypeOf(row)

	if t.Kind() == reflect.Pointer {
		t = t.Elem()
		v = v.Elem()
	}

	var setClauses []string
	namedArgs := pgx.StrictNamedArgs{}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		dbTag := field.Tag.Get("db")
		if dbTag == "" || dbTag == "-" {
			continue
		}
		setClauses = append(setClauses, dbTag+" = @"+dbTag)
		namedArgs[dbTag] = v.Field(i).Interface()
	}

	maps.Copy(namedArgs, whereArgs)

	setClauseStr := strings.Join(setClauses, ", ")

	sql := "UPDATE " + tableName + " SET " + setClauseStr
	if whereClause != "" {
		sql += " WHERE " + whereClause
	}
	sql += ";"

	return sql, namedArgs
}
