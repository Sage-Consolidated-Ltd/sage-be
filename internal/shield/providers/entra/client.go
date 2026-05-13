package entra

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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

	checkpointJSON := p.RedisClient.Get(ctx, p.CheckpointKey).Val()
	var lastCreatedTime string

	if checkpointJSON != "" {
		var checkpoint Checkpoint
		if err := json.Unmarshal([]byte(checkpointJSON), &checkpoint); err != nil {
			log.Printf("Failed to parse checkpoint: %v", err)
			lastCreatedTime = time.Now().Add(-24 * time.Hour).Format(time.RFC3339)
		}else {
			lastCreatedTime = checkpoint.LastCreatedTime
		}
	} else {
		lastCreatedTime = time.Now().Add(-24 * time.Hour).Format(time.RFC3339)
	}

	log.Printf("Polling from checkpoint: %s", lastCreatedTime)

	filter := fmt.Sprintf("createdDateTime gt %s", lastCreatedTime)
	url := fmt.Sprintf(
		"https://graph.microsoft.com/v1.0/auditLogs/signIns?$filter=%s&$top=999&$orderby=createdDateTime desc",
		filter,
	)

	var graphResp GraphResponse
	resp, err := p.RestyClient.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+token).
		SetResult(&graphResp).
		Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch sign-in logs: %w", err)
	}

	if resp.StatusCode() == http.StatusTooManyRequests {
		retryAfter := resp.Header().Get("Retry-After")
		waitSec := p.BackoffSec
		if retryAfter != "" {
			fmt.Sscanf(retryAfter, "%d", &waitSec)
		}
		log.Printf("Rate limited. Backing off %d seconds", waitSec)
		time.Sleep(time.Duration(waitSec) * time.Second)
		return []SignInEvent{}, nil
	}

	if resp.StatusCode() != http.StatusOK {

		body := resp.String()

		p.ConsecutiveErr++

		if p.ConsecutiveErr > 5 {
			log.Printf(
				"CRITICAL: Too many consecutive errors. Circuit breaker triggered.",
			)
		}

		return nil, fmt.Errorf(
			"poll failed with status %d: %s",
			resp.StatusCode(),
			body,
		)
	}

	p.ConsecutiveErr = 0
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