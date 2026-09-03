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

// @host      backend.sageconsolidated.com
// @BasePath  /api/v1
func main() {
	application, err := apiApp.New()
	if err != nil {
		log.Fatalf("failed to initialize api application: %v", err)
	}

	if err := application.Run(); err != nil {
		log.Fatalf("api application error: %v", err)
	}
}
