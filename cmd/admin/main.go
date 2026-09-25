package main

//go:generate swag init -g main.go -d ./cmd/admin,./internal/admin,./internal/shared --parseInternal -o ./docs/admin

import (
	"log"
	adminApp "sage-backend/internal/app/admin"

	_ "sage-backend/docs/admin"
)

// @title           Sage Admin API
// @version         1.0
// @description     Documentation for the Sage Super Admin / Platform Administration API.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @securityDefinitions.apikey AdminSessionAuth
// @in cookie
// @name admin_session

// @securityDefinitions.apikey AdminBearerAuth
// @in header
// @name Authorization

// @tag.name Admin Auth
// @tag.description Endpoints for platform super administrator authentication, session management, and OTP verification.

// @host      admin.sageconsolidated.com
// @BasePath  /api/v1
// @x-tagGroups [{"name":"Admin","tags":["Admin Auth"]}]
func main() {
	application, err := adminApp.New()
	if err != nil {
		log.Fatalf("failed to initialize admin application: %v", err)
	}

	if err := application.Run(); err != nil {
		log.Fatalf("admin application error: %v", err)
	}
}
