package upload_parser

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strings"

	"sage-backend/internal/shield/models"

	"github.com/google/uuid"
)

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

	fmt.Println("CSVParser.Parse: colIdx:", colIdx)
	mapper := resolveCSVMapper(colIdx)

	var out []models.ParsedLog
	for {
		row, err := r.Read()
		if err != nil {
			break
		}

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
