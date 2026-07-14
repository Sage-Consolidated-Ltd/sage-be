package upload_parser

import "database/sql"

type csvFieldMapper func(row []string, colIdx map[string]int) (sql.NullTime, string)

type CSVSourceSchema struct {
	Name    string
	Matches func(colIdx map[string]int) bool
	Mapper  csvFieldMapper
}

var csvSchemaRegistry []CSVSourceSchema

func RegisterCSVSchema(schema CSVSourceSchema) {
	csvSchemaRegistry = append(csvSchemaRegistry, schema)
}

func resolveCSVMapper(colIdx map[string]int) csvFieldMapper {
	for _, s := range csvSchemaRegistry {
		if s.Matches(colIdx) {
			return s.Mapper
		}
	}
	return genericCSVMapper
}

func genericCSVMapper(row []string, colIdx map[string]int) (sql.NullTime, string) {
	ts := ParseNullTime(FirstNonEmpty(
		GetCol(row, colIdx, "timestamp"),
		GetCol(row, colIdx, "time"),
		GetCol(row, colIdx, "date"),
	))
	level := GetCol(row, colIdx, "level")
	return ts, level
}