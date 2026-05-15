package s3

import (
	"context"
	"fmt"
	"mime/multipart"

	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
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

func (u *Uploader) UploadAvatar(ctx context.Context, file multipart.File, userID string) (string, error) {

	key := fmt.Sprintf("avatars/users/%s.png", userID)

	_, err := u.manager.UploadObject(ctx, &transfermanager.UploadObjectInput{
		Bucket: &u.client.Bucket,
		Key:    &key,
		Body:   file,
	})

	if err != nil {
		return "", err
	}

	url := fmt.Sprintf(
		"https://%s.s3.%s.amazonaws.com/%s",
		u.client.Bucket,
		u.client.Region,
		key,
	)

	return url, nil
}
