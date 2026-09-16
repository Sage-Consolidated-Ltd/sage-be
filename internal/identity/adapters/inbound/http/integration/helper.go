package integration

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

type APIResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
	Error   interface{}     `json:"error,omitempty"`
}

func performRequest(
	app *fiber.App,
	method string,
	path string,
	body any,
) (*http.Response, error) {
	return performRequestWithCookie(app, method, path, body, "")
}

func performRequestWithCookie(
	app *fiber.App,
	method string,
	path string,
	body any,
	cookieHeader string,
) (*http.Response, error) {
	var bodyReader io.Reader

	if body != nil {
		switch v := body.(type) {
		case io.Reader:
			bodyReader = v
		case []byte:
			bodyReader = bytes.NewReader(v)
		case string:
			bodyReader = strings.NewReader(v)
		default:
			jsonBytes, err := json.Marshal(v)
			if err != nil {
				return nil, err
			}
			bodyReader = bytes.NewReader(jsonBytes)
		}
	}

	req := httptest.NewRequest(method, path, bodyReader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if cookieHeader != "" {
		req.Header.Set("Cookie", cookieHeader)
	}

	return app.Test(req, -1)
}

func extractCookieHeader(resp *http.Response) string {
	cookies := resp.Header.Values("Set-Cookie")
	if len(cookies) == 0 {
		return ""
	}
	parts := strings.Split(cookies[0], ";")
	return parts[0]
}

func decodeResponse(
	t *testing.T,
	resp *http.Response,
) *APIResponse {
	t.Helper()

	var res APIResponse
	err := json.NewDecoder(resp.Body).Decode(&res)
	require.NoError(t, err)

	return &res
}
