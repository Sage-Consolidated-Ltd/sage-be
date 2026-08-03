package s3

type FileType string

const (
	Avatar   FileType = "avatar"
	Document FileType = "document"
	Temp     FileType = "temp"
)

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
