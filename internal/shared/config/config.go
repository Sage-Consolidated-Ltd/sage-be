package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
)

type BaseConfig struct {
	APP_ENV          string
	DatabaseUrl      string
	DBMAXOpenConns   int
	DBMAXIdleConns   int
	DBConnMAXLife    int
	AppEncryptionKey string
}

type APIConfig struct {
	BaseConfig
	JWTSecret          string
	PORT               string
	GomniSecurityKey   string
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

type OffenseConfig struct {
	BaseConfig
	PORT             string
	GomniSecurityKey string
}

type ShieldConfig struct {
	BaseConfig
	PORT             string
	GomniSecurityKey string
}

type VisionConfig struct {
	BaseConfig
	PORT             string
	GomniSecurityKey string
}


func requireEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("%s is required", key)
	}
	return v
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v, ok := os.LookupEnv(key); ok {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

func loadBase() BaseConfig {
	return BaseConfig{
		APP_ENV:          requireEnv("APP_ENV"),
		DatabaseUrl:      requireEnv("DATABASE_URL"),
		DBMAXOpenConns:   getEnvInt("DB_MAX_OPEN_CONNS", 25),
		DBMAXIdleConns:   getEnvInt("DB_MAX_IDLE_CONNS", 25),
		DBConnMAXLife:    getEnvInt("DB_CONN_MAX_LIFE", 300),
		AppEncryptionKey: requireEnv("APP_ENCRYPTION_KEY"),
	}
}


func SetupAPI() *APIConfig {
	_ = godotenv.Load(".env")
	return &APIConfig{
		BaseConfig:         loadBase(),
		JWTSecret:          requireEnv("JWT_SECRET"),
		PORT:               getEnv("PORT", "3333"),
		GomniSecurityKey:   requireEnv("GOMNI_SECURITY_KEY"),
		GoogleClientId:     requireEnv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: requireEnv("GOOGLE_CLIENT_SECRET"),
		GoogleRedirectUrl:  requireEnv("GOOGLE_REDIRECT_URL"),
		GitHubClientId:     requireEnv("GITHUB_CLIENT_ID"),
		GitHubClientSecret: requireEnv("GITHUB_CLIENT_SECRET"),
		GitHubRedirectUrl:  requireEnv("GITHUB_REDIRECT_URL"),
		AzureClientId:      requireEnv("AZURE_CLIENT_ID"),
		AzureClientSecret:  requireEnv("AZURE_CLIENT_SECRET"),
		AzureRedirectUrl:   requireEnv("AZURE_REDIRECT_URL"),
	}
}

func SetupOffense() *OffenseConfig {
	_ = godotenv.Load(".env")
	return &OffenseConfig{
		BaseConfig:       loadBase(),
		PORT:             getEnv("OFFENSE_PORT", "3334"),
	}
}

func SetupShield() *ShieldConfig {
	_ = godotenv.Load(".env")
	return &ShieldConfig{
		BaseConfig:       loadBase(),
		PORT:             getEnv("SHIELD_PORT", "3335"),
	}
}

func SetupVision() *VisionConfig {
	_ = godotenv.Load(".env")
	return &VisionConfig{
		BaseConfig:       loadBase(),
		PORT:             getEnv("VISION_PORT", "3336"),
	}
}