package ai_detector

import (
	"context"
	"fmt"
	"io"
	"sage-backend/internal/shield/models"
	"strings"

	"github.com/go-resty/resty/v2"
)

type AIDetectorClient struct {
	BaseURL     string
	Token       string
	restyClient *resty.Client
}

func NewAIDetectorClient(baseURL, token string, restyClient *resty.Client) *AIDetectorClient {
	trimmedBaseURL := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	restyClient.SetBaseURL(trimmedBaseURL)

	if strings.TrimSpace(token) != "" {
		restyClient.SetAuthToken(strings.TrimSpace(token))
	}

	return &AIDetectorClient{
		BaseURL:     trimmedBaseURL,
		Token:       strings.TrimSpace(token),
		restyClient: restyClient,
	}
}

func (a *AIDetectorClient) Health(ctx context.Context) (*models.CheckHealthResponse, error) {
	var result models.CheckHealthResponse

	resp, err := a.restyClient.R().
		SetContext(ctx).
		SetResult(&result).
		Get("/health")
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("detector API error: %s", resp.String())
	}

	return &result, nil
}
func (a *AIDetectorClient) DetectFileThreats(
	ctx context.Context,
	file io.Reader,
	filename string,
) (*models.AnalysisResult, error) {

	var result models.AnalysisResult

	resp, err := a.restyClient.R().
		SetContext(ctx).
		SetFileReader("file", filename, file).
		SetResult(&result).
		Post("/v1/detect?approach=A&threshold=0.5")

	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, fmt.Errorf("detector API error: %s", resp.String())
	}

	return &result, nil
}
