package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"go.uber.org/zap/zapcore"
)

type BaseConfig struct {
	APP_ENV          string
	DatabaseUrl      string
	DBMAXOpenConns   int
	DBMAXIdleConns   int
	DBConnMAXLife    int
	AppEncryptionKey string
	RedisDbUrl       string
	RedisHost        string
	RedisPort        int
	RedisPassword    string
	RedisDB          int
	LogLevel         Level
	ResendApiKey     string
	ResendFromEmail  string
	FrontendBaseURL  string
	S3Region         string
	S3Bucket         string
	S3AccessKey      string
	S3SecretKey      string
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
	PORT                string
	GomniSecurityKey    string
	DetectorAIBaseURL   string
	DetectorAIAuthToken string
}

type VisionConfig struct {
	BaseConfig
	PORT             string
	GomniSecurityKey string
}

type Level = zapcore.Level

const (
	DebugLevel = zapcore.DebugLevel
	InfoLevel  = zapcore.InfoLevel
	WarnLevel  = zapcore.WarnLevel
	ErrorLevel = zapcore.ErrorLevel
	FatalLevel = zapcore.FatalLevel
)

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

func levelFromEnv() Level {
	switch strings.ToLower(os.Getenv("LOG_LEVEL")) {
	case "debug":
		return DebugLevel
	case "warn", "warning":
		return WarnLevel
	case "error":
		return ErrorLevel
	case "fatal":
		return FatalLevel
	default:
		return InfoLevel
	}
}

func loadBase() BaseConfig {
	return BaseConfig{
		APP_ENV:          requireEnv("APP_ENV"),
		DatabaseUrl:      requireEnv("DATABASE_URL"),
		DBMAXOpenConns:   getEnvInt("DB_MAX_OPEN_CONNS", 25),
		DBMAXIdleConns:   getEnvInt("DB_MAX_IDLE_CONNS", 25),
		DBConnMAXLife:    getEnvInt("DB_CONN_MAX_LIFE", 300),
		AppEncryptionKey: requireEnv("APP_ENCRYPTION_KEY"),
		RedisDbUrl:       requireEnv("REDIS_DB_URL"),
		RedisHost:        requireEnv("REDIS_HOST"),
		RedisPort:        getEnvInt("REDIS_PORT", 6379),
		RedisPassword:    getEnv("REDIS_PASSWORD", ""),
		RedisDB:          getEnvInt("REDIS_DB", 0),
		LogLevel:         levelFromEnv(),
		ResendApiKey:     requireEnv("RESEND_API_KEY"),
		ResendFromEmail:  requireEnv("RESEND_FROM_EMAIL"),
		FrontendBaseURL:  requireEnv("FRONTEND_BASE_URL"),
		S3Region:         requireEnv("S3_REGION"),
		S3AccessKey:      requireEnv("AWS_ACCESS_KEY_ID"),
		S3SecretKey:      requireEnv("AWS_SECRET_ACCESS_KEY"),
		S3Bucket:         requireEnv("S3_BUCKET"),
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
		BaseConfig: loadBase(),
		PORT:       getEnv("OFFENSE_PORT", "3334"),
	}
}

func SetupShield() *ShieldConfig {
	_ = godotenv.Load(".env")
	return &ShieldConfig{
		BaseConfig:          loadBase(),
		PORT:                getEnv("SHIELD_PORT", "3335"),
		DetectorAIBaseURL:   requireEnv("DETECTOR_AI_BASE_URL"),
		DetectorAIAuthToken: requireEnv("DETECTOR_AI_AUTH_TOKEN"),
	}
}

func SetupVision() *VisionConfig {
	_ = godotenv.Load(".env")
	return &VisionConfig{
		BaseConfig: loadBase(),
		PORT:       getEnv("VISION_PORT", "3336"),
	}
}
