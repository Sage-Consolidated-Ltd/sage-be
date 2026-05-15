package okta

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"time"
)

func (p *OktaProvider) GetSignInEvents(ctx context.Context, since time.Time) ([]OktaSignInEvent, error) {
	// Okta filter: after=<timestamp>
	filter := fmt.Sprintf("after=%s", since.Format(time.RFC3339))
	url := fmt.Sprintf("%s/api/v1/logs?%s&limit=200", p.RestyClient.DisableWarn, p.Domain, filter)

	resp, err := p.RestyClient.R().
		SetContext(ctx).
		SetHeader("Accept", "application/json").
		SetHeader("Authorization", fmt.Sprintf("SSWS %s", p.ApiToken)).
		Get(url)

	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	// Handle rate limiting (429)
	if resp.StatusCode() == 429 {
		return nil, fmt.Errorf("rate limited (429), retry after: %s", resp.Header().Get("Retry-After"))
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode(), resp.String())
	}

	var events []OktaSignInEvent
	if err := json.Unmarshal(resp.Body(), &events); err != nil {
		return nil, fmt.Errorf("decode failed: %w", err)
	}

	return events, nil
}

func (p *OktaProvider) PollAuditLogs(
	ctx context.Context,
) ([]OktaSignInEvent, error) {

	// Get checkpoint
	var lastCreatedTime string

	if p.Checkpoint == nil || p.Checkpoint.LastCheckpoint == nil {
		lastCreatedTime = time.Now().
			Add(-24 * time.Hour).
			Format(time.RFC3339)

	} else {
		lastCreatedTime = *p.Checkpoint.LastCheckpoint
	}

	log.Printf(
		"Polling Okta logs from checkpoint: %s",
		lastCreatedTime,
	)

	// Build query params
	params := url.Values{}

	params.Add(
		"since",
		lastCreatedTime,
	)

	params.Add(
		"limit",
		"1000",
	)

	params.Add(
		"filter",
		`eventType eq "user.session.start"`,
	)

	endpoint := fmt.Sprintf(
		"%s/api/v1/logs?%s",
		p.Domain,
		params.Encode(),
	)

	var events []OktaSignInEvent

	resp, err := p.RestyClient.R().
		SetContext(ctx).
		SetHeader(
			"Authorization",
			"SSWS "+p.ApiToken,
		).
		SetHeader(
			"Accept",
			"application/json",
		).
		SetResult(&events).
		Get(endpoint)

	if err != nil {
		log.Printf(
			"Error fetching Okta logs: %v",
			err,
		)

		return nil,
			fmt.Errorf(
				"failed to fetch okta logs: %w",
				err,
			)
	}

	// Handle rate limiting
	if resp.StatusCode() == 429 {
		retryAfter :=
			resp.Header().
				Get("Retry-After")

		return nil, fmt.Errorf("provider rate limited (retry-after=%s)", retryAfter)
	}

	// Handle failures
	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf(
			"okta poll failed with status %d: %s",
			resp.StatusCode(),
			resp.String(),
		)
	}
	log.Printf(
		"Fetched %d Okta events",
		len(events),
	)

	return events, nil
}

func (p *OktaProvider) HealthCheck(ctx context.Context) error {
	url := fmt.Sprintf("%s/api/v1/user/me", p.Domain)

	resp, err := p.RestyClient.R().
		SetContext(ctx).
		SetHeader("Accept", "application/json").
		SetHeader("Authorization", fmt.Sprintf("SSWS %s", p.ApiToken)).
		Get(url)

	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	if resp.StatusCode() != 200 {
		return fmt.Errorf("health check returned %d", resp.StatusCode)
	}

	return nil
}
