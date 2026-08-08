package upload_parser

import (
	"fmt"

	"sage-backend/internal/shield/domain"
)

type WindowsEventLogParser struct{}

func (WindowsEventLogParser) Parse(raw []byte) ([]domain.ParsedLog, error) {
	return nil, fmt.Errorf("WindowsEventLogParser.Parse: not implemented — plug in evtx/text parser")
}
