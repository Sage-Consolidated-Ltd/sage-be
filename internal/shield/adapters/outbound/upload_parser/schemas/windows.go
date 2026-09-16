package schemas

import (
	"database/sql"

	"sage-backend/internal/shield/adapters/outbound/upload_parser"
)

var eventTypeToLevel = map[string]string{
	"1":  "error",
	"2":  "warning",
	"4":  "info",
	"8":  "info",
	"16": "warning",
}

func windowsSecurityCSVMapper(row []string, colIdx map[string]int) (sql.NullTime, string) {
	ts := upload_parser.ParseNullTime(upload_parser.GetCol(row, colIdx, "timegenerated"))

	level := ""
	if et := upload_parser.GetCol(row, colIdx, "eventtype"); et != "" {
		level = eventTypeToLevel[et]
	}
	return ts, level
}

func init() {
	upload_parser.RegisterCSVSchema(upload_parser.CSVSourceSchema{
		Name: "windows_security",
		Matches: func(colIdx map[string]int) bool {
			_, hasTG := colIdx["timegenerated"]
			_, hasET := colIdx["eventtype"]
			_, hasEID := colIdx["eventid"]
			return hasTG && hasET && hasEID
		},
		Mapper: windowsSecurityCSVMapper,
	})
}
