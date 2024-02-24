package models

import (
	"database/sql"
	"fmt"
	"strings"
)

type Row interface {
	Err() error
	Scan(dest ...any) error
}

func columnsNoID(columns string) string {
	if len(columns) == 1 {
		return ""
	}

	c := strings.Split(columns, " ")
	return strings.Join(c[1:], " ")
}

func values(columnsS string) string {
	columns := strings.Split(columnsS, ", ")
	var vals []string
	for _, c := range columns {
		if c != "" {
			vals = append(vals, "?")
		}
	}

	return strings.Join(vals, ", ")
}

func set(columnsS string) string {
	columns := strings.Split(columnsS, ", ")
	var vals []string
	for _, c := range columns {
		if c != "" {
			vals = append(vals, fmt.Sprintf("%s=?", c))
		}
	}

	return strings.Join(vals, ", ")
}

func toInt64(i sql.NullInt64) *int64 {
	if i.Valid {
		return &i.Int64
	} else {
		return nil
	}
}

func toString(i sql.NullString) *string {
	if i.Valid {
		return &i.String
	} else {
		return nil
	}
}
