package upload_parser

import (
	"fmt"

	"sage-backend/internal/shield/models"
)

type Parser interface {
	Parse(raw []byte) ([]models.ParsedLog, error)
}

var parserRegistry = map[models.DetectedType]Parser{}

func RegisterParser(t models.DetectedType, p Parser) {
	parserRegistry[t] = p
}

func ParserFor(t models.DetectedType) (Parser, error) {
	p, ok := parserRegistry[t]
	fmt.Println("ParserFor: t:", t, "p:", p, "ok:", ok)
	if !ok {
		return nil, fmt.Errorf("no parser for detected type %q", t)
	}
	return p, nil
}
