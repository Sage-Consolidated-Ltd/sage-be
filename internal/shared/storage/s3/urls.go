package s3

import "fmt"

func (c *S3Client) PublicURL(key string) string {
	return fmt.Sprintf(
		"https://%s.s3.%s.amazonaws.com/%s",
		c.cfg.Bucket,
		c.cfg.Region,
		key,
	)
}
