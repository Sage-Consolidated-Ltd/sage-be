package okta

import "time"

type OktaSignInEvent struct {
	EventID   string    `json:"eventId"`
	Published time.Time `json:"published"`
	Actor     struct {
		ID          string `json:"id"`
		DisplayName string `json:"displayName"`
		Login       string `json:"login"`
	} `json:"actor"`
	Client struct {
		IPAddress string `json:"ipAddress"`
		UserAgent string `json:"userAgent"`
	} `json:"client"`
	Outcome struct {
		Result string `json:"result"`
		Reason string `json:"reason"`
	} `json:"outcome"`
	Target []struct {
		ID          string `json:"id"`
		DisplayName string `json:"displayName"`
		Type        string `json:"type"`
	} `json:"target"`
}
