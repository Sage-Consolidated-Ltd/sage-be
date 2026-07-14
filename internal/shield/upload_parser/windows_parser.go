package upload_parser

import (
	"fmt"

	"sage-backend/internal/shield/models"
)

type WindowsEventLogParser struct{}

func (WindowsEventLogParser) Parse(raw []byte) ([]models.ParsedLog, error) {
	return nil, fmt.Errorf("WindowsEventLogParser.Parse: not implemented — plug in evtx/text parser")
}