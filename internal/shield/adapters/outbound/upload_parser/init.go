package upload_parser

import "sage-backend/internal/shield/domain"

func init() {
	RegisterParser(domain.DetectedCSVStructured, CSVParser{})
	RegisterParser(domain.DetectedXLSXStructured, XLSXParser{})
	RegisterParser(domain.DetectedLinuxSyslog, LinuxSyslogParser{})
	RegisterParser(domain.DetectedWindowsEventLog, WindowsEventLogParser{})
}
