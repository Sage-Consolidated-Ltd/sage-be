package upload_parser

import (
	"bufio"
	"bytes"
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"time"

	"sage-backend/internal/shield/domain"

	"github.com/google/uuid"
)

type LinuxSyslogParser struct{}

var syslogExtract = regexp.MustCompile(
	`^([A-Z][a-z]{2}\s+\d{1,2}\s+\d{2}:\d{2}:\d{2})\s+(\S+)\s+(\S+):\s+(.*)$`,
)

func (LinuxSyslogParser) Parse(raw []byte) ([]domain.ParsedLog, error) {
	scanner := bufio.NewScanner(bytes.NewReader(raw))
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var out []domain.ParsedLog

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		matches := syslogExtract.FindStringSubmatch(line)

		if len(matches) == 5 {
			out = append(out, domain.ParsedLog{
				ID:        uuid.New(),
				Timestamp: ToNullTime(parseSyslogTime(matches[1])),
				Message:   matches[4],
				RawJSON: map[string]any{
					"host":    matches[2],
					"program": matches[3],
					"raw":     line,
				},
			})
			continue
		}

		out = append(out, domain.ParsedLog{
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

func parseSyslogTime(s string) time.Time {
	t, err := time.Parse("Jan 2 15:04:05", s)
	if err != nil {
		return time.Time{}
	}

	now := time.Now()
	return time.Date(now.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), 0, time.UTC)
}
