package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
)

type Config struct {
	APP_ENV        string
	DatabaseUrl    string
	DBMAXOpenConns int
	DBMAXIdleConns int
	DBConnMAXLife  int
	JWTSecret      string
	PORT           string
}

func requireEnv(value string) string {
	if os.Getenv(value) == "" {
		log.Fatalf("%s is required", value)
	}
	return os.Getenv(value)
}

func getEnv(value, defaultStr string) string {
	if os.Getenv(value) == "" {
		return defaultStr
	}
	return os.Getenv(value)
}

func getEnvInt(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return fallback
}

func Setup() *Config {
	_ = godotenv.Load(".env")
	app_env := requireEnv("APP_ENV")

	return &Config{
		APP_ENV:        app_env,
		DatabaseUrl:    requireEnv("DATABASE_URL"),
		JWTSecret:      requireEnv("JWT_SECRET"),
		DBMAXOpenConns: getEnvInt("DB_MAX_OPEN_CONNS", 25),
		DBMAXIdleConns: getEnvInt("DB_MAX_IDLE_CONNS", 25),
		DBConnMAXLife:  getEnvInt("DB_CONN_MAX_LIFE", 300),
		PORT:           getEnv("PORT", "3333"),
	}
}
