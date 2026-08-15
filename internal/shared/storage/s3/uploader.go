package s3

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"sage-backend/internal/shared/middlewares"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var allowedExts = map[string]FileInfo{
	".json": {Class: ClassJSON, ContentType: "application/json", S3Prefix: "json/", Allowed: true},
	".csv":  {Class: ClassCSV, ContentType: "text/csv", S3Prefix: "csv/", Allowed: true},
	".xml":  {Class: ClassXML, ContentType: "application/xml", S3Prefix: "xml/", Allowed: true},
	".gz":   {Class: ClassCompressed, ContentType: "application/gzip", S3Prefix: "compressed/", Allowed: true},
	".zip":  {Class: ClassCompressed, ContentType: "application/zip", S3Prefix: "compressed/", Allowed: true},
	".tar":  {Class: ClassCompressed, ContentType: "application/x-tar", S3Prefix: "compressed/", Allowed: true},
	".pcap": {Class: ClassPCAP, ContentType: "application/vnd.tcpdump.pcap", S3Prefix: "pcap/", Allowed: true},
	".log":  {Class: ClassSyslog, ContentType: "text/plain", S3Prefix: "syslog/", Allowed: true},
	".evt":  {Class: ClassWindowsEvent, ContentType: "application/octet-stream", S3Prefix: "windows_event/", Allowed: true},
	".evtx": {Class: ClassWindowsEvent, ContentType: "application/octet-stream", S3Prefix: "windows_event/", Allowed: true},
	".txt":  {Class: ClassSyslog, ContentType: "text/plain", S3Prefix: "syslog/", Allowed: true},
	".yml":  {Class: ClassSyslog, ContentType: "text/yaml", S3Prefix: "syslog/", Allowed: true},
	".yaml": {Class: ClassSyslog, ContentType: "text/yaml", S3Prefix: "syslog/", Allowed: true},
}

type Uploader struct {
	client          *Client
	manager         *transfermanager.Client
	logUploadPolicy *UploadPolicy
}

func NewUploader(client *Client) *Uploader {
	maxSize := int64(100 * 1024 * 1024)
	if client.cfg.MaxFileSizeMB > 0 {
		maxSize = client.cfg.MaxFileSizeMB * 1024 * 1024
	}
	return &Uploader{
		client: client,
		manager: transfermanager.New(client.S3, func(u *transfermanager.Options) {
			u.PartSizeBytes = 5 * 1024 * 1024
			u.Concurrency = 5
		}),
		logUploadPolicy: &UploadPolicy{
			AllowedExt: allowedExts,
			MaxSize:    maxSize,
		},
	}
}

func (u *Uploader) UploadAvatar(ctx context.Context, file multipart.File, userID string, mimeType string) (string, string, error) {
	magicBytes := make([]byte, 512)
	n, _ := file.Read(magicBytes)
	if n == 0 {
		return "", "", errors.New("empty file")
	}
	file.Seek(0, 0)

	detectedType := http.DetectContentType(magicBytes[:n])
	if detectedType != mimeType || !middlewares.AvatarMimeTypes[detectedType] {
		return "", "", fmt.Errorf("file content mismatch or invalid type")
	}

	ext, err := determineImageExtension(mimeType)
	if err != nil {
		return "", "", err
	}

	key := fmt.Sprintf("avatars/users/%s%s", userID, ext)

	_, err = u.manager.UploadObject(ctx, &transfermanager.UploadObjectInput{
		Bucket: &u.client.Bucket,
		Key:    &key,
		Body:   file,
		Metadata: map[string]string{
			"Content-Disposition": "inline",
		},
	})

	if err != nil {
		return "", "", err
	}

	presignerClient := s3.NewPresignClient(u.client.S3)

	presignedRequest, err := presignerClient.PresignGetObject(ctx,
		&s3.GetObjectInput{
			Bucket: aws.String(u.client.Bucket),
			Key:    aws.String(key),
		},
		func(opts *s3.PresignOptions) {
			opts.Expires = time.Duration(24 * time.Hour)
		},
	)
	if err != nil {
		return "", "", err
	}

	return presignedRequest.URL, key, nil
}

func (u *Uploader) UploadLogo(ctx context.Context, file multipart.File, orgID string, logoType string, mimeType string) (string, string, error) {
	magicBytes := make([]byte, 512)
	n, _ := file.Read(magicBytes)
	if n == 0 {
		return "", "", errors.New("empty file")
	}
	file.Seek(0, 0)

	detectedType := http.DetectContentType(magicBytes[:n])
	if detectedType != mimeType || !middlewares.AvatarMimeTypes[detectedType] {
		return "", "", fmt.Errorf("file content mismatch or invalid type")
	}

	ext, err := determineImageExtension(mimeType)
	if err != nil {
		return "", "", err
	}

	key := fmt.Sprintf("branding/logos/%s_%s%s", orgID, logoType, ext)

	_, err = u.manager.UploadObject(ctx, &transfermanager.UploadObjectInput{
		Bucket: &u.client.Bucket,
		Key:    &key,
		Body:   file,
		Metadata: map[string]string{
			"Content-Disposition": "inline",
		},
	})

	if err != nil {
		return "", "", err
	}

	presignerClient := s3.NewPresignClient(u.client.S3)

	presignedRequest, err := presignerClient.PresignGetObject(ctx,
		&s3.GetObjectInput{
			Bucket: aws.String(u.client.Bucket),
			Key:    aws.String(key),
		},
		func(opts *s3.PresignOptions) {
			opts.Expires = time.Duration(24 * time.Hour)
		},
	)
	if err != nil {
		return "", "", err
	}

	return presignedRequest.URL, key, nil
}

func (u *Uploader) GenerateSignedURL(ctx context.Context, key string) (string, error) {
	presignerClient := s3.NewPresignClient(u.client.S3)

	presignRequest, err := presignerClient.PresignGetObject(ctx,
		&s3.GetObjectInput{
			Bucket:                     &u.client.Bucket,
			Key:                        aws.String(key),
			ResponseContentDisposition: aws.String("inline"),
		},
		func(opts *s3.PresignOptions) {
			opts.Expires = time.Duration(24 * time.Hour)
		},
	)
	if err != nil {
		return "", err
	}

	return presignRequest.URL, nil
}

func (u *Uploader) DownloadObject(ctx context.Context, key string) ([]byte, error) {
	result, err := u.client.S3.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(u.client.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("get object: %w", err)
	}
	defer result.Body.Close()

	return io.ReadAll(result.Body)
}

func (u *Uploader) HeadObject(ctx context.Context, key string) (*s3.HeadObjectOutput, error) {
	return u.client.S3.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(u.client.Bucket),
		Key:    aws.String(key),
	})
}

func (u *Uploader) ValidateETag(s3ETag, clientETag string) bool {
	clean := func(e string) string {
		if len(e) >= 2 && e[0] == '"' && e[len(e)-1] == '"' {
			return e[1 : len(e)-1]
		}
		return e
	}
	return clean(s3ETag) == clean(clientETag)
}

func (u *Uploader) PresignUploadPost(
	ctx context.Context,
	rc middlewares.RequestContext,
	filename string,
	sizeBytes int64,
) (*PresignedPost, string, *time.Time, error) {
	info, err := u.logUploadPolicy.Validate(filename, sizeBytes)
	if err != nil {
		return nil, "", nil, err
	}

	ext := strings.ToLower(filepath.Ext(filename))
	key := fmt.Sprintf(
		"uploads/pending/%s/%s/%s-%d%s",
		info.Class,
		rc.UserID,
		rc.RequestID,
		time.Now().UnixNano(),
		ext,
	)

	expiryMin := u.client.cfg.PresignExpiry
	if expiryMin <= 0 {
		expiryMin = 1440
	}
	expiration := time.Now().Add(time.Duration(expiryMin) * time.Minute)
	now := time.Now().UTC()
	creds := buildCredential(u.client.cfg, now)

	maxSize := u.logUploadPolicy.MaxSize
	if maxSize <= 0 {
		maxSize = 100 * 1024 * 1024
	}

	conditions := []interface{}{
		map[string]string{"bucket": u.client.Bucket},
		map[string]string{"key": key},
		[]interface{}{"starts-with", "$Content-Type", ""},
		[]interface{}{"content-length-range", 0, maxSize},
		map[string]string{"x-amz-algorithm": "AWS4-HMAC-SHA256"},
		map[string]string{"x-amz-credential": creds},
		map[string]string{"x-amz-date": now.Format("20060102T150405Z")},
		[]interface{}{"starts-with", "$x-amz-meta-", ""},
		[]interface{}{"starts-with", "$x-amz-meta-expected-class", ""},
		[]interface{}{"starts-with", "$x-amz-meta-original-name", ""},
		[]interface{}{"starts-with", "$x-amz-meta-upload-source", ""},
	}
	if u.client.cfg.SessionToken != "" {
		conditions = append(conditions, map[string]string{"x-amz-security-token": u.client.cfg.SessionToken})
	}

	policy := map[string]interface{}{
		"expiration": expiration.UTC().Format(time.RFC3339),
		"conditions": conditions,
	}

	policyBytes, err := json.Marshal(policy)
	if err != nil {
		return nil, "", nil, fmt.Errorf("marshal policy: %w", err)
	}
	policyBase64 := base64.StdEncoding.EncodeToString(policyBytes)

	fields := PresignedPostFields{
		Key:                   key,
		ContentType:           info.ContentType,
		Policy:                policyBase64,
		XAmzAlgorithm:         "AWS4-HMAC-SHA256",
		XAmzCredential:        creds,
		XAmzDate:              now.Format("20060102T150405Z"),
		XAmzSignature:         signPolicy(policyBase64, u.client.cfg, now),
		XAmzSecurityToken:     u.client.cfg.SessionToken,
		XAmzMetaExpectedClass: string(info.Class),
		XAmzMetaOriginalName:  filename,
		XAmzMetaUploadSource:  "siem",
	}

	return &PresignedPost{
		URL:    fmt.Sprintf("https://%s.s3.%s.amazonaws.com", u.client.Bucket, u.client.Region),
		Fields: fields,
	}, key, &expiration, nil
}

func determineImageExtension(mimeType string) (string, error) {
	var ext string
	switch mimeType {
	case "image/jpeg":
		ext = ".jpg"
	case "image/png":
		ext = ".png"
	case "image/webp":
		ext = ".webp"
	default:
		return "", errors.New("unsupported image type")
	}

	return ext, nil
}
