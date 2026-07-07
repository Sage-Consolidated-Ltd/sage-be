package s3

import (
	"bytes"
	"path/filepath"
	"strings"
)

var magicMap = []magicSig{
	{[]byte{0x45, 0x6c, 0x66, 0x46, 0x69, 0x6c, 0x65, 0x00}, ClassWindowsEvent},

	{[]byte{0xd4, 0xc3, 0xb2, 0xa1}, ClassPCAP},
	{[]byte{0xa1, 0xb2, 0xc3, 0xd4}, ClassPCAP},
	{[]byte{0x0a, 0x0d, 0x0d, 0x0a}, ClassPCAP},

	{[]byte{0x1f, 0x8b}, ClassCompressed},
	{[]byte{0x50, 0x4b, 0x03, 0x04}, ClassCompressed},

	{[]byte{0x7b}, ClassJSON},
	{[]byte{0x5b}, ClassJSON},
	{[]byte{0x3c}, ClassXML},
}

var extMap = map[string]FileInfo{
	// Windows Logs
	".evtx": {Class: ClassWindowsEvent, ContentType: "application/octet-stream", S3Prefix: "windows-events", Allowed: true},
	".evt": {Class: ClassWindowsEvent, ContentType: "application/octet-stream", S3Prefix: "windows-events", Allowed: true},
	".etl": {Class: ClassWindowsEvent, ContentType: "application/octet-stream", S3Prefix: "windows-events", Allowed: true},

	// Network captures
	".pcap": {Class: ClassPCAP, ContentType: "application/vnd.tcpdump.pcap", S3Prefix: "pcap", Allowed: true},
	".pcapng": {Class: ClassPCAP, ContentType: "application/vnd.tcpdump.pcap", S3Prefix: "pcap", Allowed: true},
	".har": {Class: ClassPCAP, ContentType: "application/json", S3Prefix: "pcap", Allowed: true},

	// Log formats
	".log": {Class: ClassSyslog, ContentType: "text/plain", S3Prefix: "syslog", Allowed: true},
	".syslog": {Class: ClassSyslog, ContentType: "text/plain", S3Prefix: "syslog", Allowed: true},
	".cef": {Class: ClassSyslog, ContentType: "text/plain", S3Prefix: "syslog", Allowed: true},

	// Structured formats
	".json": {
		Class: ClassJSON,
		ContentType: "application/json",
		S3Prefix: "json-logs",
		Allowed: true,
	},
	".jsonl": {
		Class: ClassJSON,
		ContentType: "application/x-ndjson",
		S3Prefix: "json-logs",
		Allowed: true,
	},
	".ndjson": {
		Class: ClassJSON,
		ContentType: "application/x-ndjson",
		S3Prefix: "json-logs",
		Allowed: true,
	},
	".xml": {
		Class: ClassXML,
		ContentType: "application/xml",
		S3Prefix: "xml-logs",
		Allowed: true,
	},
	".csv": {
		Class: ClassCSV,
		ContentType: "text/csv",
		S3Prefix: "csv-logs",
		Allowed: true,
	},
	".tsv": {
		Class: ClassCSV,
		ContentType: "text/tab-separated-values",
		S3Prefix: "csv-logs",
		Allowed: true,
	},

	// Compression
	".gz": {
		Class: ClassCompressed,
		ContentType: "application/gzip",
		S3Prefix: "compressed",
		Allowed: true,
	},
	".gzip": {
		Class: ClassCompressed,
		ContentType: "application/gzip",
		S3Prefix: "compressed",
		Allowed: true,
	},
	".zip": {
		Class: ClassCompressed,
		ContentType: "application/zip",
		S3Prefix: "compressed",
		Allowed: true,
	},
	".bz2": {
		Class: ClassCompressed,
		ContentType: "application/x-bzip2",
		S3Prefix: "compressed",
		Allowed: true,
	},
	".xz": {
		Class: ClassCompressed,
		ContentType: "application/x-xz",
		S3Prefix: "compressed",
		Allowed: true,
	},
	".zst": {
		Class: ClassCompressed,
		ContentType: "application/zstd",
		S3Prefix: "compressed",
		Allowed: true,
	},
	".7z": {
		Class: ClassCompressed,
		ContentType: "application/x-7z-compressed",
		S3Prefix: "compressed",
		Allowed: true,
	},
	".tar": {
		Class: ClassCompressed,
		ContentType: "application/x-tar",
		S3Prefix: "compressed",
		Allowed: true,
	},
	".tgz": {
		Class: ClassCompressed,
		ContentType: "application/gzip",
		S3Prefix: "compressed",
		Allowed: true,
	},

	// Threat intel
	".stix": {
		Class: ClassThreatIntel,
		ContentType: "application/stix+json",
		S3Prefix: "threat-intel",
		Allowed: true,
	},
	".ioc": {
		Class: ClassThreatIntel,
		ContentType: "application/xml",
		S3Prefix: "threat-intel",
		Allowed: true,
	},
	".yar": {
		Class: ClassThreatIntel,
		ContentType: "text/plain",
		S3Prefix: "threat-intel",
		Allowed: true,
	},
	".yara": {
		Class: ClassThreatIntel,
		ContentType: "text/plain",
		S3Prefix: "threat-intel",
		Allowed: true,
	},
}

func Classify(filename string) FileInfo {
	name := strings.ToLower(filename)

	// Try compound extensions first
	compoundExts := []string{
		".json.gz",
		".jsonl.gz",
		".ndjson.gz",
		".log.gz",
		".pcap.gz",
		".evtx.gz",
		".tar.gz",
		".tgz",
	}

	for _, ext := range compoundExts {
		if strings.HasSuffix(name, ext) {
			switch ext {
			case ".json.gz", ".jsonl.gz", ".ndjson.gz":
				return FileInfo{
					Class:       ClassJSON,
					ContentType: "application/gzip",
					S3Prefix:    "json-logs",
					Allowed:     true,
				}

			case ".log.gz":
				return FileInfo{
					Class:       ClassSyslog,
					ContentType: "application/gzip",
					S3Prefix:    "syslog",
					Allowed:     true,
				}

			case ".pcap.gz":
				return FileInfo{
					Class:       ClassPCAP,
					ContentType: "application/gzip",
					S3Prefix:    "pcap",
					Allowed:     true,
				}

			case ".evtx.gz":
				return FileInfo{
					Class:       ClassWindowsEvent,
					ContentType: "application/gzip",
					S3Prefix:    "windows-events",
					Allowed:     true,
				}

			case ".tar.gz", ".tgz":
				return FileInfo{
					Class:       ClassCompressed,
					ContentType: "application/gzip",
					S3Prefix:    "compressed",
					Allowed:     true,
				}
			}
		}
	}

	// Fallback to regular extension
	ext := filepath.Ext(name)

	if info, ok := extMap[ext]; ok {
		return info
	}

	return FileInfo{
		Class:       ClassUnknown,
		ContentType: "application/octet-stream",
		S3Prefix:    "unknown",
		Allowed:     false,
	}
}

func ClassifyFromMagic(firstBytes []byte, claimed FileInfo) bool {
	for _, m := range magicMap {
		if len(firstBytes) < len(m.sig) {
			continue
		}

		if bytes.Equal(firstBytes[:len(m.sig)], m.sig) {
			return m.class == claimed.Class
		}
	}

	// IMPORTANT: do NOT blindly allow unknown binaries
	return false
}