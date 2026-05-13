package entra

import (
	"context"
	"fmt"
	"log"
	"time"
)

func (p *EntraProvider) getToken(ctx context.Context) (string, error) {
	p.TokenMutex.RLock()
	if p.AccessToken != "" && time.Now().Before(p.TokenExpiresAt.Add(-5*time.Minute)) {
		defer p.TokenMutex.RUnlock()
		return p.AccessToken, nil
	}

	p.TokenMutex.RUnlock()
	log.Println("Refreshing Entra access token...")

	tokenURL := fmt.Sprintf(
		"https://login.microsoftonline.com/%s/oauth2/v2.0/token",
		p.TenantID,
	)

	body := fmt.Sprintf(
		"grant_type=client_credentials&client_id=%s&client_secret=%s&scope=https://graph.microsoft.com/.default",
		p.ClientID,
		p.ClientSecret,
	)

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}

	resp, err := p.RestyClient.R().
		SetResult(&tokenResp).
		SetContext(ctx).
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetBody(body).
		Post(tokenURL)
	if err != nil {
		return "", fmt.Errorf("failed to refresh Entra access token: %w", err)
	}

	if !resp.IsSuccess() {
		return "", fmt.Errorf("failed to refresh Entra access token: %s", resp.Status())
	}

	p.TokenMutex.Lock()
	p.AccessToken = tokenResp.AccessToken

	p.TokenExpiresAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	p.TokenMutex.Unlock()

	return tokenResp.AccessToken, nil
}
func (p *EntraProvider) IsExpiringSoon() bool {
	p.TokenMutex.RLock()
	defer p.TokenMutex.RUnlock()

	if p.AccessToken == "" {
		return true
	}
 
	return time.Now().After(p.TokenExpiresAt.Add(-5*time.Minute))
}
