package s3

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func NewClient(cfg S3Config) (*S3Client, error) {
	awsConfig, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(cfg.Region),
	)
	if err != nil {
		return nil, err
	}

	c := s3.NewFromConfig(awsConfig)
	return &S3Client{
		S3:     c,
		Presign: s3.NewPresignClient(c),
		cfg: cfg,
	}, nil
}
