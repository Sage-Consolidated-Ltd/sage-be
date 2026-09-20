package upload_parser

import (
	"fmt"

	"sage-backend/internal/shield/domain"
)

type Parser interface {
	Parse(raw []byte) ([]domain.ParsedLog, error)
}

var parserRegistry = map[domain.DetectedType]Parser{}

func RegisterParser(t domain.DetectedType, p Parser) {
	parserRegistry[t] = p
}

func ParserFor(t domain.DetectedType) (Parser, error) {
	p, ok := parserRegistry[t]
	if !ok {
		return nil, fmt.Errorf("no parser for detected type %q", t)
	}
	return p, nil
}
