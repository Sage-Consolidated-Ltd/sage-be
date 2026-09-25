package main

//go:generate swag init -g main.go -d ./cmd/api,./internal/identity,./internal/organization,./internal/shared --parseInternal -o ./docs/users

import (
	"log"
	apiApp "sage-backend/internal/app/api"

	_ "sage-backend/docs/users"
)

// @title           Sage API
// @version         1.0
// @description     Documentation for the Sage API (Users).
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @securityDefinitions.apikey SessionAuth
// @in cookie
// @name session_id

// @tag.name Auth
// @tag.description Endpoints for user registration, authentication, 2FA, and password recovery.
// @tag.name User Profile
// @tag.description Endpoints for user profile, sessions, activity logs, and account preferences.
// @tag.name Company
// @tag.description Endpoints for company setup, invitations, and industry selection.
// @tag.name Organization
// @tag.description Endpoints for organization management, members, branding, and custom RBAC roles.
// @tag.name Organization Dashboard
// @tag.description Endpoints for organization posture score, compliance, and real-time dashboard telemetry.

// @host      backend.sageconsolidated.com
// @BasePath  /api/v1
// @x-tagGroups [{"name":"Identity","tags":["Auth","User Profile"]},{"name":"Organization","tags":["Company","Organization","Organization Dashboard"]}]
func main() {
	application, err := apiApp.New()
	if err != nil {
		log.Fatalf("failed to initialize api application: %v", err)
	}

	if err := application.Run(); err != nil {
		log.Fatalf("api application error: %v", err)
	}
}
