package config

import (
    "golang.org/x/oauth2"
    "golang.org/x/oauth2/google"
    "golang.org/x/oauth2/github"
    "golang.org/x/oauth2/microsoft"
)

type OAuthConfig struct {
    GoogleClientId     string
    GoogleClientSecret string
    GoogleRedirectUrl  string
    
    GitHubClientId     string
    GitHubClientSecret string
    GitHubRedirectUrl  string
    
    AzureClientId      string
    AzureClientSecret  string
    AzureRedirectUrl   string
}

// NewOAuthConfig is your new constructor
func NewOAuthConfig(
    googleId, googleSecret, googleRedirect string,
    githubId, githubSecret, githubRedirect string,
    azureId, azureSecret, azureRedirect string,
) *OAuthConfig {
    return &OAuthConfig{
        GoogleClientId:     googleId,
        GoogleClientSecret: googleSecret,
        GoogleRedirectUrl:  googleRedirect,
        GitHubClientId:     githubId,
        GitHubClientSecret: githubSecret,
        GitHubRedirectUrl:  githubRedirect,
        AzureClientId:      azureId,
        AzureClientSecret:  azureSecret,
        AzureRedirectUrl:   azureRedirect,
    }
}

// GetConfig returns the provider-specific oauth2.Config
func (c *OAuthConfig) GetConfig(provider string) *oauth2.Config {
    switch provider {
    case "google":
        return &oauth2.Config{
            ClientID:     c.GoogleClientId,
            ClientSecret: c.GoogleClientSecret,
            RedirectURL:  c.GoogleRedirectUrl,
            Endpoint:     google.Endpoint,
            Scopes:       []string{"https://www.googleapis.com/auth/userinfo.profile", "https://www.googleapis.com/auth/userinfo.email"},
        }
    case "github":
        return &oauth2.Config{
            ClientID:     c.GitHubClientId,
            ClientSecret: c.GitHubClientSecret,
            RedirectURL:  c.GitHubRedirectUrl,
            Endpoint:     github.Endpoint,
            Scopes:       []string{"read:user", "user:email"},
        }
    case "azure":
        return &oauth2.Config{
            ClientID:     c.AzureClientId,
            ClientSecret: c.AzureClientSecret,
            RedirectURL:  c.AzureRedirectUrl,
            Endpoint:     microsoft.AzureADEndpoint("common"),
            Scopes:       []string{"https://graph.microsoft.com/User.Read"},
        }
    }
    return nil
}