package config

import (
	"github.com/stretchr/gomniauth"
	"github.com/stretchr/gomniauth/providers/google"
	"github.com/stretchr/gomniauth/providers/github"
)

type ConfigGomniAuth struct {
	securityKey string
	googleClientId string
	googleClientSecret string
	redirectUrl string
	gitHubClientId string
	gitHubClientSecret string
	gitHubRedirectUrl string
}

func NewConfigGomniAuth(
	securityKey string,
	googleClientId string,
	googleClientSecret string,
	redirectUrl string,
	gitHubClientId string,
	gitHubClientSecret string,
	gitHubRedirectUrl string,
) *ConfigGomniAuth {
	return &ConfigGomniAuth{
		securityKey: securityKey,
		googleClientId: googleClientId,
		googleClientSecret: googleClientSecret,
		redirectUrl: redirectUrl,
		gitHubClientId: gitHubClientId,
		gitHubClientSecret: gitHubClientSecret,
		gitHubRedirectUrl: gitHubRedirectUrl,
	}
}

func (c *ConfigGomniAuth) InitGomniauth() {
	gomniauth.SetSecurityKey(c.securityKey)
	gomniauth.WithProviders(
		google.New(
			c.googleClientId,
			c.googleClientSecret,
			c.redirectUrl,
		),
		github.New(
			c.gitHubClientId,
			c.gitHubClientSecret,
			c.gitHubRedirectUrl,
		),
	)
}