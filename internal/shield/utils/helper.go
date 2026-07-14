package utils

import (
	"bufio"
	"bytes"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"sage-backend/internal/shield/models"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
)

type Parser interface {
	Parse(raw []byte) ([]models.ParsedLog, error)
}

// ---------------- DETECTION ----------------

var syslogPattern = regexp.MustCompile(
	`^[A-Z][a-z]{2}\s+\d{1,2}\s+\d{2}:\d{2}:\d{2}\s+\S+\s+\S+(\[\d+\])?:`,
)

func DetectType(fileName string, content []byte) models.DetectedType {
	ext := strings.ToLower(filepath.Ext(fileName))
	text := string(content)

	switch ext {
	case ".csv":
		return models.DetectedCSVStructured
	case ".xlsx":
		return models.DetectedXLSXStructured
	}

	if looksLikeWindowsEventLog(text) {
		return models.DetectedWindowsEventLog
	}
	if looksLikeSyslog(text) {
		return models.DetectedLinuxSyslog
	}

	return models.DetectedUnknown
}

func looksLikeWindowsEventLog(text string) bool {
	return strings.Contains(text, "Event ID:") &&
		strings.Contains(text, "Level:") &&
		strings.Contains(text, "Date:")
}

func looksLikeSyslog(text string) bool {
	lines := strings.Split(text, "\n")
	matched := 0
	checked := 0

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		checked++
		if syslogPattern.MatchString(line) {
			matched++
		}
		if checked >= 5 {
			break
		}
	}
	return checked > 0 && matched*2 >= checked
}

// ---------------- PARSER REGISTRY ----------------

var parserRegistry = map[models.DetectedType]Parser{}

func RegisterParser(t models.DetectedType, p Parser) {
	parserRegistry[t] = p
}

func ParserFor(t models.DetectedType) (Parser, error) {
	p, ok := parserRegistry[t]
	if !ok {
		return nil, fmt.Errorf("no parser for detected type %q", t)
	}
	return p, nil
}

// ---------------- CSV PARSER ----------------

type CSVParser struct{}

func (CSVParser) Parse(raw []byte) ([]models.ParsedLog, error) {
	r := csv.NewReader(bytes.NewReader(raw))

	header, err := r.Read()
	if err != nil {
		return nil, fmt.Errorf("csv header: %w", err)
	}

	colIdx := make(map[string]int)
	for i, h := range header {
		colIdx[strings.ToLower(strings.TrimSpace(h))] = i
	}

	var out []models.ParsedLog

	for {
		row, err := r.Read()
		if err != nil {
			break
		}

		out = append(out, models.ParsedLog{
			ID: uuid.New(),
			Timestamp: parseNullTime(firstNonEmpty(
				getCol(row, colIdx, "timestamp"),
				getCol(row, colIdx, "time"),
				getCol(row, colIdx, "date"),
			)),
			Level:   getCol(row, colIdx, "level"),
			Message: getCol(row, colIdx, "message"),
			RawJSON: rowToMap(header, row),
		})
	}

	return out, nil
}

// ---------------- XLSX PARSER ----------------

type XLSXParser struct{}

func (XLSXParser) Parse(raw []byte) ([]models.ParsedLog, error) {
	f, err := excelize.OpenReader(bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}

	sheet := f.GetSheetName(0)
	rows, err := f.GetRows(sheet)
	if err != nil {
		return nil, err
	}

	if len(rows) == 0 {
		return nil, fmt.Errorf("empty sheet")
	}

	header := rows[0]
	colIdx := make(map[string]int)
	for i, h := range header {
		colIdx[strings.ToLower(strings.TrimSpace(h))] = i
	}

	var out []models.ParsedLog

	for _, row := range rows[1:] {
		out = append(out, models.ParsedLog{
			ID: uuid.New(),
			Timestamp: parseNullTime(firstNonEmpty(
				getCol(row, colIdx, "timestamp"),
				getCol(row, colIdx, "time"),
				getCol(row, colIdx, "date"),
			)),
			Level:   getCol(row, colIdx, "level"),
			Message: getCol(row, colIdx, "message"),
			RawJSON: rowToMap(header, row),
		})
	}

	return out, nil
}

// ---------------- SYSLOG PARSER ----------------

type LinuxSyslogParser struct{}

var syslogExtract = regexp.MustCompile(
	`^([A-Z][a-z]{2}\s+\d{1,2}\s+\d{2}:\d{2}:\d{2})\s+(\S+)\s+(\S+):\s+(.*)$`,
)

func (LinuxSyslogParser) Parse(raw []byte) ([]models.ParsedLog, error) {
	scanner := bufio.NewScanner(bytes.NewReader(raw))
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var out []models.ParsedLog

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		matches := syslogExtract.FindStringSubmatch(line)

		if len(matches) == 5 {
			out = append(out, models.ParsedLog{
				ID:        uuid.New(),
				Timestamp: toNullTime(parseSyslogTime(matches[1])),
				Message:   matches[4],
				RawJSON: map[string]any{
					"host":    matches[2],
					"program": matches[3],
					"raw":     line,
				},
			})
			continue
		}

		out = append(out, models.ParsedLog{
			ID:        uuid.New(),
			Timestamp: sql.NullTime{Valid: false},
			Message:   line,
			RawJSON: map[string]any{
				"raw_line": line,
			},
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("syslog scan: %w", err)
	}

	return out, nil
}

// ---------------- WINDOWS PARSER (STUB) ----------------

type WindowsEventLogParser struct{}

var eventTypeToLevel = map[string]string{
	"1":  "error",   // Error
	"2":  "warning", // Warning
	"4":  "info",    // Information
	"8":  "info",    // Success Audit
	"16": "warning", // Failure Audit
}

func (WindowsEventLogParser) Parse(raw []byte) ([]models.ParsedLog, error) {
	scanner := bufio.NewScanner(bytes.NewReader(raw))
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var out []models.ParsedLog

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var fields map[string]any
		if err := json.Unmarshal([]byte(line), &fields); err != nil {
			// skip malformed lines rather than failing the whole file
			continue
		}

		message, _ := fields["Message"].(string)

		var ts sql.NullTime
		if tg, ok := fields["TimeGenerated"].(string); ok {
			if t, err := time.Parse(time.RFC3339, tg); err == nil {
				ts = sql.NullTime{Time: t, Valid: true}
			}
		}

		level := ""
		if et, ok := fields["EventType"].(string); ok {
			level = eventTypeToLevel[et]
		}

		out = append(out, models.ParsedLog{
			ID:        uuid.New(),
			Timestamp: ts,
			Level:     level,
			Message:   message,
			RawJSON:   fields,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("windows event log scan: %w", err)
	}

	return out, nil
}

// ---------------- HELPERS ----------------

func getCol(row []string, idx map[string]int, name string) string {
	if i, ok := idx[name]; ok && i < len(row) {
		return row[i]
	}
	return ""
}

func rowToMap(header, row []string) map[string]any {
	m := make(map[string]any, len(header))
	for i, h := range header {
		if i < len(row) {
			m[h] = row[i]
		}
	}
	return m
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func parseTimeOrZero(s string) time.Time {
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

func parseNullTime(s string) sql.NullTime {
	t := parseTimeOrZero(s)

	if t.IsZero() {
		return sql.NullTime{Valid: false}
	}

	return sql.NullTime{
		Time:  t,
		Valid: true,
	}
}

func toNullTime(t time.Time) sql.NullTime {
	if t.IsZero() {
		return sql.NullTime{Valid: false}
	}

	return sql.NullTime{
		Time:  t,
		Valid: true,
	}
}

func parseSyslogTime(s string) time.Time {
	t, err := time.Parse("Jan 2 15:04:05", s)
	if err != nil {
		return time.Time{}
	}

	now := time.Now()

	return time.Date(
		now.Year(),
		t.Month(),
		t.Day(),
		t.Hour(),
		t.Minute(),
		t.Second(),
		0,
		time.UTC,
	)
}

// ---------------- INIT ----------------

func init() {
	RegisterParser(models.DetectedCSVStructured, CSVParser{})
	RegisterParser(models.DetectedXLSXStructured, XLSXParser{})
	RegisterParser(models.DetectedLinuxSyslog, LinuxSyslogParser{})
	RegisterParser(models.DetectedWindowsEventLog, WindowsEventLogParser{})
}
