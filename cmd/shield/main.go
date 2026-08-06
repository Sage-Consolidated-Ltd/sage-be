package main

import (
	"log"
	shieldApp "sage-backend/internal/app/shield"

	_ "sage-backend/docs/shield"
)

// @title           Sage API (Shield)
// @version         1.0
// @description     Documentation for the Sage API (Shield).
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
	application, err := shieldApp.New()
	if err != nil {
		log.Fatalf("failed to initialize shield application: %v", err)
	}

	if err := application.Run(); err != nil {
		log.Fatalf("shield application error: %v", err)
	}
}
