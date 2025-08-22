package utils

import (
	"fmt"
	"reflect"
	"strings"
)

// goTypeToSQL maps Go types to their corresponding SQL types.
func goTypeToSQL(t reflect.Type) string {
	switch t.Kind() {
	case reflect.String:
		return "TEXT"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return "INTEGER"
	case reflect.Bool:
		return "BOOLEAN"
	case reflect.Float32, reflect.Float64:
		return "REAL"
	}
	// Special case for time.Time
	if t.String() == "time.Time" {
		return "TIMESTAMP"
	}
	return "TEXT" // Default to TEXT for other types
}

// GenerateCreateTableSQL generates a CREATE TABLE SQL statement from a Go struct.
func GenerateCreateTableSQL(s interface{}) (string, error) {
	t := reflect.TypeOf(s)
	if t.Kind() != reflect.Struct {
		return "", fmt.Errorf("expected a struct")
	}

	tableName := strings.ToLower(t.Name())
	var columns []string
	var primaryKeys []string

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		dbTag := field.Tag.Get("db")
		if dbTag == "" {
			continue // Skip fields without a db tag
		}

		columnName := dbTag
		sqlType := goTypeToSQL(field.Type)
		columns = append(columns, fmt.Sprintf(`"%s" %s`, columnName, sqlType))

		if pkTag := field.Tag.Get("pk"); pkTag == "true" {
			primaryKeys = append(primaryKeys, fmt.Sprintf(`"%s"`, columnName))
		}
	}

	if len(columns) == 0 {
		return "", fmt.Errorf("no fields with db tags found in struct")
	}

	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS "%s" (%s`, tableName, strings.Join(columns, ", "))
	if len(primaryKeys) > 0 {
		query += fmt.Sprintf(", PRIMARY KEY (%s)", strings.Join(primaryKeys, ", "))
	}
	query += ");"

	return query, nil
}
