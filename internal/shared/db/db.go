package db

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"log"
	"sage-backend/internal/shared/config"
	"time"

	_ "github.com/lib/pq"

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

type JSONMap map[string]interface{}

func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan JSON")
	}

	return json.Unmarshal(bytes, j)
}
func (j JSONMap) Value() (driver.Value, error) {
	return json.Marshal(j)
}

type JSONMapSlice []JSONMap

func (j *JSONMapSlice) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan JSONMapSlice")
	}

	return json.Unmarshal(bytes, j)
}
func (j JSONMapSlice) Value() (driver.Value, error) {
	return json.Marshal(j)
}

type GenericJSON[T any] struct {
	Data T
}

func (j *GenericJSON[T]) Scan(value any) error {
	if value == nil {
		j.Data = *new(T)
		return nil
	}

	var data []byte

	switch v := value.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return fmt.Errorf("cannot scan %T into JSON", value)
	}

	return json.Unmarshal(data, &j.Data)
}

func (j GenericJSON[T]) Value() (driver.Value, error) {
	return json.Marshal(j.Data)
}