package s3

import (
	"mime/multipart"
	"time"
)

type FileType string

const (
	Avatar            FileType = "avatar"
	Document          FileType = "document"
	Temp              FileType = "temp"
	ClassWindowsEvent FileType = "windows_event"
	ClassPCAP         FileType = "pcap"
	ClassSyslog       FileType = "syslog"
	ClassJSON         FileType = "json"
	ClassCSV          FileType = "csv"
	ClassXML          FileType = "xml"
	ClassCompressed   FileType = "compressed"
	ClassThreatIntel  FileType = "threat_intel"
	ClassUnknown      FileType = "unknown"
)

type S3Config struct {
	Region          string
	Bucket          string
	PresignExpiry   int // minutes
	MaxFileSizeMB   int64
	AccessKeyID     string
	SecretAccessKey string
}

type PresignResult struct {
	URL       string
	Headers   map[string]string
	ExpiresAt time.Time
}

type FileInfo struct {
	Class       FileType
	ContentType string
	S3Prefix    string
	Allowed     bool
}

type UploadInput struct {
	File      multipart.File
	Filename  string
	MimeType  string
	Folder    string
	UserID    string
	RequestID string
	MaxSize   int64
}

type PresignedPostFields struct {
	Key                   string `json:"key"`
	ContentType           string `json:"Content-Type"`
	Policy                string `json:"policy"`
	XAmzAlgorithm         string `json:"x-amz-algorithm"`
	XAmzCredential        string `json:"x-amz-credential"`
	XAmzDate              string `json:"x-amz-date"`
	XAmzSignature         string `json:"x-amz-signature"`
	XAmzMetaExpectedClass string `json:"x-amz-meta-expected-class"`
	XAmzMetaOriginalName  string `json:"x-amz-meta-original-name"`
	XAmzMetaUploadSource  string `json:"x-amz-meta-upload-source"`
}

type PresignedPost struct {
	URL    string              `json:"url"`
	Fields PresignedPostFields `json:"fields"`
}
