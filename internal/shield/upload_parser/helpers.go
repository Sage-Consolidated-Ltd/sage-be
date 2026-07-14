package upload_parser

import (
	"database/sql"
	"strings"
	"time"
)

func GetCol(row []string, idx map[string]int, name string) string {
	if i, ok := idx[name]; ok && i < len(row) {
		return row[i]
	}
	return ""
}

func RowToMap(header, row []string) map[string]any {
	m := make(map[string]any, len(header))
	for i, h := range header {
		if i < len(row) {
			m[h] = row[i]
		}
	}
	return m
}

func FirstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func ParseTimeOrZero(s string) time.Time {
	for _, layout := range []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
	} {
		if t, err := time.Parse(layout, s); err == nil {
			return t
		}
	}
	return time.Time{}
}

func ParseNullTime(s string) sql.NullTime {
	t := ParseTimeOrZero(s)
	if t.IsZero() {
		return sql.NullTime{Valid: false}
	}
	return sql.NullTime{Time: t, Valid: true}
}

func ToNullTime(t time.Time) sql.NullTime {
	if t.IsZero() {
		return sql.NullTime{Valid: false}
	}
	return sql.NullTime{Time: t, Valid: true}
}