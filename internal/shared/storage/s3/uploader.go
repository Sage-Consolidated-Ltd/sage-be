package s3

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"sage-backend/internal/shared/middlewares"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Uploader struct {
	client  *Client
	manager *transfermanager.Client
}

func NewUploader(client *Client) *Uploader {
	return &Uploader{
		client: client,
		manager: transfermanager.New(client.S3, func(u *transfermanager.Options) {
			u.PartSizeBytes = 5 * 1024 * 1024
			u.Concurrency = 5
		}),
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
