package db

import (
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"sage-backend/internal/shared/config"
	"time"

	"github.com/jmoiron/sqlx"
)

type DB struct {
	*sqlx.DB
}

func ConnectDB(cfg *config.BaseConfig) (*DB, error) {
	db, err := sqlx.Connect("postgres", cfg.DatabaseUrl)
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to db: %w", err)
	}
	if pingerr := db.Ping(); pingerr != nil {
		return nil, fmt.Errorf("Failed to ping db: %w", pingerr)
	}

	db.SetMaxOpenConns(cfg.DBMAXOpenConns)
	db.SetMaxIdleConns(cfg.DBMAXIdleConns)
	db.SetConnMaxLifetime(time.Duration(cfg.DBConnMAXLife) * time.Second)

	log.Println("Database Connected")

	return &DB{db}, nil
}
