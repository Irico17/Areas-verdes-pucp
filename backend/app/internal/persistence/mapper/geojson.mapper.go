// Package mapper maps persistence models and database rows to domain entities and DTOs.
package mapper

import (
	"database/sql"
	"strings"
)

// GeomJSON safely parses geometry JSON from database scan.
func GeomJSON(v sql.NullString) []byte {
	if !v.Valid || strings.TrimSpace(v.String) == "" {
		return []byte("null")
	}
	return []byte(v.String)
}

// NullStr converts a sql.NullString to *string. Returns nil if invalid or empty.
func NullStr(v sql.NullString) *string {
	if !v.Valid {
		return nil
	}
	s := strings.TrimSpace(v.String)
	if s == "" {
		return nil
	}
	return &s
}

// NullFloat converts a sql.NullFloat64 to *float64. Returns nil if invalid.
func NullFloat(v sql.NullFloat64) *float64 {
	if !v.Valid {
		return nil
	}
	f := v.Float64
	return &f
}
