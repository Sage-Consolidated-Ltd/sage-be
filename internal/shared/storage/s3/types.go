package s3

type FileType string

const (
	Avatar   FileType = "avatar"
	Document FileType = "document"
	Temp     FileType = "temp"
)
