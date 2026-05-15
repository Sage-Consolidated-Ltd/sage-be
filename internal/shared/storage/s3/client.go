package s3

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Client struct {
	S3     *s3.Client
	Bucket string
	Region string
}

func NewClient(ctx context.Context, bucket, region string) (*Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
	)
	if err != nil {
		return nil, err
	}

	return &Client{
		S3:     s3.NewFromConfig(cfg),
		Bucket: bucket,
		Region: region,
	}, nil
}
