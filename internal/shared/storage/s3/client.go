package s3

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Client struct {
	S3      *s3.Client
	Bucket  string
	Region  string
	cfg     S3Config
}

func NewClient(ctx context.Context, bucket, region string) (*Client, error) {
	return NewClientWithConfig(ctx, S3Config{
		Bucket:        bucket,
		Region:        region,
		PresignExpiry: 1440,
		MaxFileSizeMB: 100,
	})
}

func NewClientWithConfig(ctx context.Context, s3Cfg S3Config) (*Client, error) {
	var optFns []func(*config.LoadOptions) error
	if s3Cfg.Region != "" {
		optFns = append(optFns, config.WithRegion(s3Cfg.Region))
	}
	if s3Cfg.AccessKeyID != "" && s3Cfg.SecretAccessKey != "" {
		optFns = append(optFns, config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(s3Cfg.AccessKeyID, s3Cfg.SecretAccessKey, ""),
		))
	}

	cfg, err := config.LoadDefaultConfig(ctx, optFns...)
	if err != nil {
		return nil, err
	}

	// Resolve credentials from loaded config provider chain if not passed directly in s3Cfg
	if s3Cfg.AccessKeyID == "" || s3Cfg.SecretAccessKey == "" {
		if creds, err := cfg.Credentials.Retrieve(ctx); err == nil {
			if s3Cfg.AccessKeyID == "" {
				s3Cfg.AccessKeyID = creds.AccessKeyID
			}
			if s3Cfg.SecretAccessKey == "" {
				s3Cfg.SecretAccessKey = creds.SecretAccessKey
			}
		}
	}

	if s3Cfg.PresignExpiry <= 0 {
		s3Cfg.PresignExpiry = 1440
	}
	if s3Cfg.MaxFileSizeMB <= 0 {
		s3Cfg.MaxFileSizeMB = 100
	}

	return &Client{
		S3:     s3.NewFromConfig(cfg),
		Bucket: s3Cfg.Bucket,
		Region: s3Cfg.Region,
		cfg:    s3Cfg,
	}, nil
}
