package entra

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"
)

func (p *EntraProvider) GetSignInLogs(ctx context.Context, filter string) (*GraphResponse, error) {
	if filter == "" {
		log.Println("Warning: No filter provided for sign-in logs query")
	}

	url := fmt.Sprintf(
		"%s/auditLogs/signIns?$filter=%s&$top=999&$orderby=createdDateTime desc",
		p.BaseUrl,
		filter,
	)

	token, err := p.getToken(ctx)
	log.Println("Entra token: ", token)
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	resp, err := p.RestyClient.R().
		SetContext(ctx).
		SetHeader("Accept", "application/json").
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", token)).
		Get(url)

	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode() == 429 {
		return nil, fmt.Errorf("rate limited (429), retry after: %s", resp.Header().Get("Retry-After"))
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode(), resp.String())
	}

	var graphResp GraphResponse
	if err := json.Unmarshal(resp.Body(), &graphResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &graphResp, nil
}
func (p *EntraProvider) PollAuditLogs(ctx context.Context) ([]SignInEvent, error) {
	token, err := p.getToken(ctx)
	if err != nil {
		return nil, err
	}

	// checkpointJSON := p.RedisClient.Get(ctx, p.CheckpointKey).Val()
	var lastCreatedTime string

	if p.Checkpoint == nil || p.Checkpoint.LastCheckpoint == nil {
		lastCreatedTime = time.Now().
			Add(-24 * time.Hour).
			Format(time.RFC3339)

	} else {
		lastCreatedTime = *p.Checkpoint.LastCheckpoint
	}

	log.Printf("Polling from checkpoint: %s", lastCreatedTime)

	filter := fmt.Sprintf(
		"createdDateTime gt %s",
		lastCreatedTime,
	)

	params := url.Values{}
	params.Add("$filter", filter)
	params.Add("$top", "999")
	params.Add("$orderby", "createdDateTime desc")

	url := fmt.Sprintf(
		"https://graph.microsoft.com/v1.0/auditLogs/signIns?%s",
		params.Encode(),
	)

	var graphResp GraphResponse
	resp, err := p.RestyClient.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+token).
		SetResult(&graphResp).
		Get(url)
	if err != nil {
		log.Printf("Error fetching sign in logs: %s", err.Error())
		return nil, fmt.Errorf("failed to fetch sign-in logs: %w", err)
	}

	if resp.StatusCode() == http.StatusTooManyRequests {
		retryAfter := resp.Header().Get("Retry-After")

		return nil, fmt.Errorf(
			"entra rate limited (retry-after=%s)",
			retryAfter,
		)
	}

	if resp.StatusCode() != http.StatusOK {
		body := resp.String()

		return nil, fmt.Errorf(
			"poll failed with status %d: %s",
			resp.StatusCode(),
			body,
		)
	}

	log.Printf("Received Response: %v", graphResp.Value)

	log.Printf(
		"Fetched %d sign-in events",
		len(graphResp.Value),
	)

	return graphResp.Value, nil
}
func (p *EntraProvider) HealthCheck(ctx context.Context) error {
	url := fmt.Sprintf("%s/me", p.BaseUrl)

	token, err := p.getToken(ctx)
	if err != nil {
		return fmt.Errorf("failed to get token: %w", err)
	}

	resp, err := p.RestyClient.R().
		SetContext(ctx).
		SetHeader("Accept", "application/json").
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", token)).
		Get(url)

	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode() != 200 {
		return fmt.Errorf(
			"health check failed with status %d: %s",
			resp.StatusCode(),
			resp.String(),
		)
	}

	return nil
}
