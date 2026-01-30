package dbastidor

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"

	"github.com/wrapped-owls/testereiro/puppetest/pkg/strnormalizer"
)

func ExecuteSeedStruct(db *sql.DB, item any) error {
	val := reflect.ValueOf(item)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return fmt.Errorf("expected struct or pointer to struct, got %s", val.Kind())
	}

	typ := val.Type()
	tableName := strnormalizer.ToSnakeCase(typ.Name()) + "s" // Simple pluralization standard

	var columns []string
	var values []any
	var placeholders []string

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		dbTag := field.Tag.Get("db")
		if dbTag == "" || dbTag == "-" {
			continue
		}

		columns = append(columns, dbTag)
		values = append(values, val.Field(i).Interface())
		placeholders = append(placeholders, "?")
	}

	if len(columns) == 0 {
		return fmt.Errorf("no database fields found for struct %s", typ.Name())
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		tableName,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	_, err := db.Exec(query, values...)
	if err != nil {
		return fmt.Errorf("failed to insert into %s: %w", tableName, err)
	}

	return nil
}
