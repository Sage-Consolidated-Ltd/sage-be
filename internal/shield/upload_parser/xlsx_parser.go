package upload_parser

import (
	"bytes"
	"fmt"
	"strings"

	"sage-backend/internal/shield/models"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
)

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

	mapper := resolveCSVMapper(colIdx)

	var out []models.ParsedLog
	for _, row := range rows[1:] {
		timestamp, level := mapper(row, colIdx)

		out = append(out, models.ParsedLog{
			ID:        uuid.New(),
			Timestamp: timestamp,
			Level:     level,
			Message:   GetCol(row, colIdx, "message"),
			RawJSON:   RowToMap(header, row),
		})
	}

	return out, nil
}