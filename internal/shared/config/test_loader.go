package config

import (
	"fmt"
	"strings"

	"github.com/joho/godotenv"
)

type TestConfigOption func(*APIConfig)

// SetupTestConfig constructs a dedicated configuration for integration tests.
// It defaults to isolated test resources (sage_db_test database and Redis DB index 2)
// and enforces strict safety checks to prevent tests from executing against development/production resources.
func SetupTestConfig(opts ...TestConfigOption) *APIConfig {
	_ = godotenv.Overload(".env.test")

	dbURL := getEnv("TEST_DATABASE_URL", getEnv("DATABASE_URL", "postgres://peterpaul:sage_dev_password@localhost:5432/sage_db_test?sslmode=disable"))
	redisURL := getEnv("TEST_REDIS_DB_URL", getEnv("REDIS_DB_URL", "redis://localhost:6379/2"))

	cfg := &APIConfig{
		BaseConfig: BaseConfig{
			APP_ENV:          "test",
			DatabaseUrl:      dbURL,
			DBMAXOpenConns:   10,
			DBMAXIdleConns:   5,
			DBConnMAXLife:    300,
			AppEncryptionKey: getEnv("APP_ENCRYPTION_KEY", "test-encryption-key-32-bytes-long"),
			RedisDbUrl:       redisURL,
			LogLevel:         DebugLevel,
			ResendApiKey:     "test",
			ResendFromEmail:  "test@example.com",
			FrontendBaseURL:  "http://localhost:3000",
			S3Bucket:         "sage-uploads-test",
			S3Region:         "us-east-1",
		},
		JWTSecret:          getEnv("JWT_SECRET", "test-jwt-secret-key-32-bytes-long"),
		PORT:               getEnv("PORT", "4001"),
		GomniSecurityKey:   getEnv("GOMNI_SECURITY_KEY", "test-security-key"),
		GoogleClientId:     "mock-google-client-id",
		GoogleClientSecret: "mock-google-client-secret",
		GoogleRedirectUrl:  "http://localhost:4000/api/v1/auth/google/callback",
		GitHubClientId:     "mock-github-client-id",
		GitHubClientSecret: "mock-github-client-secret",
		GitHubRedirectUrl:  "http://localhost:4000/api/v1/auth/github/callback",
		AzureClientId:      "mock-azure-client-id",
		AzureClientSecret:  "mock-azure-client-secret",
		AzureRedirectUrl:   "http://localhost:4000/api/v1/auth/azure/callback",
		S3Region:           "us-east-1",
		S3AccessKey:        "mock-aws-access-key",
		S3SecretKey:        "mock-aws-secret-key",
		S3Bucket:           "sage-uploads-test",
	}

	for _, opt := range opts {
		opt(cfg)
	}

	if err := validateTestSafety(cfg); err != nil {
		panic(fmt.Sprintf("CRITICAL TEST SAFETY VIOLATION: %v", err))
	}

	return cfg
}

// validateTestSafety enforces that test configuration explicitly points to isolated test resources.
func validateTestSafety(cfg *APIConfig) error {
	env := strings.ToLower(cfg.APP_ENV)
	if env != "test" && env != "testing" {
		return fmt.Errorf("refusing to run integration tests with APP_ENV=%q (must be 'test' or 'testing')", cfg.APP_ENV)
	}

	if !strings.Contains(strings.ToLower(cfg.DatabaseUrl), "test") {
		return fmt.Errorf("refusing to run integration tests against database URL %q (database name must contain 'test', e.g. 'sage_db_test')", cfg.DatabaseUrl)
	}

	if strings.HasSuffix(cfg.RedisDbUrl, "/0") || strings.Contains(cfg.RedisDbUrl, "/0?") {
		return fmt.Errorf("refusing to run integration tests against Redis DB 0 (must use dedicated non-zero test DB index, e.g. DB 2)")
	}

	return nil
}
