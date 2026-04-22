package main

import (
	"log"
	"sage-backend/internal/shared/config"
	"sage-backend/internal/shared/db"
)

func main() {
	cfg := config.SetupShield()
	db, err := db.ConnectDB(&cfg.BaseConfig)
	if err != nil {
		log.Fatalf("Error connecting to db: %s", err)
	}
	defer db.Close()

	
}