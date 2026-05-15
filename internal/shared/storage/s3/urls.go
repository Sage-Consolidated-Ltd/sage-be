package s3

import "fmt"

func (c *Client) PublicURL(key string) string {
	return fmt.Sprintf(
		"https://%s.s3.%s.amazonaws.com/%s",
		c.Bucket,
		c.Region,
		key,
	)
}
