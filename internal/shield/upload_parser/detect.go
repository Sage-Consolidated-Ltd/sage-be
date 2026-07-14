package upload_parser

import (
	"path/filepath"
	"regexp"
	"strings"

	"sage-backend/internal/shield/models"
)

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
