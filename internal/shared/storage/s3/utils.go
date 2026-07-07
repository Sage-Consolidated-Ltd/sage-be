package s3

import (
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"fmt"
	"net/http"
	"time"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	"github.com/go-resty/resty/v2"
)

func hmacSHA256(key []byte, data string) []byte {
	h := hmac.New(sha256.New, key)
	h.Write([]byte(data))
	return h.Sum(nil)
}

// deriveSigningKey produces the AWS v4 signing key for a given date + region + service.
func deriveSigningKey(secret, date, region, service string) []byte {
	kDate := hmacSHA256([]byte("AWS4"+secret), date)
	kRegion := hmacSHA256(kDate, region)
	kService := hmacSHA256(kRegion, service)
	kSigning := hmacSHA256(kService, "aws4_request")
	return kSigning
}

// buildCredential returns the x-amz-credential string.
func buildCredential(cfg S3Config, now time.Time) string {
	fmt.Printf("DEBUG buildCredential: AccessKeyID=%q Region=%q\n", cfg.AccessKeyID, cfg.Region)
	return fmt.Sprintf("%s/%s/%s/s3/aws4_request",
		cfg.AccessKeyID,
		now.Format("20060102"),
		cfg.Region,
	)
}

// signPolicy signs the base64-encoded policy document and returns a hex signature.
func signPolicy(policyBase64 string, cfg S3Config, now time.Time) string {
	key := deriveSigningKey(cfg.SecretAccessKey, now.Format("20060102"), cfg.Region, "s3")
	sig := hmacSHA256(key, policyBase64)
	return fmt.Sprintf("%x", sig)
}

func NotFoundError(err error) bool {
	if err == nil {
		return false
	}
	var re *http.Response
	_ = re
	return err != nil && (isHTTPStatus(err, 404) || isNoSuchKey(err))
}

func isHTTPStatus(err error, code int) bool {
	var restyErr *resty.ResponseError

	if errors.As(err, &restyErr) {
		return restyErr.Response.StatusCode() == code
	}
	return false
}

func isNoSuchKey(err error) bool {
	if err == nil {
		return false
	}

	var nsk *types.NoSuchKey
	return errors.As(err, &nsk)
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
