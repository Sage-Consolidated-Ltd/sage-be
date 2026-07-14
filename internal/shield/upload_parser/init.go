package upload_parser

import "sage-backend/internal/shield/models"

func init() {
	RegisterParser(models.DetectedCSVStructured, CSVParser{})
	RegisterParser(models.DetectedXLSXStructured, XLSXParser{})
	RegisterParser(models.DetectedLinuxSyslog, LinuxSyslogParser{})
	RegisterParser(models.DetectedWindowsEventLog, WindowsEventLogParser{})
}