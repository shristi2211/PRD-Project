package config

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	DatabaseURL       string
	RedisURL          string
	Port              string
	CORSOrigins       string
	PrivateKeyPath    string
	PublicKeyPath     string
	AccessTokenExpiry time.Duration
	RefreshTokenExpiry time.Duration
}

// Load reads environment variables and returns a validated Config.
// It fails fast if any required variable is missing.
func Load() (*Config, error) {
	// Load .env file (ignore error if not found — production uses real env vars)
	_ = godotenv.Load()

	cfg := &Config{
		DatabaseURL:    os.Getenv("DATABASE_URL"),
		RedisURL:       getEnvOrDefault("REDIS_URL", "redis://localhost:6379"),
		Port:           getEnvOrDefault("PORT", "8080"),
		CORSOrigins:    getEnvOrDefault("CORS_ORIGINS", "http://localhost:5173"),
		PrivateKeyPath: getEnvOrDefault("JWT_PRIVATE_KEY_PATH", "keys/private.pem"),
		PublicKeyPath:  getEnvOrDefault("JWT_PUBLIC_KEY_PATH", "keys/public.pem"),
	}

	// Parse durations
	accessExp, err := time.ParseDuration(getEnvOrDefault("ACCESS_TOKEN_EXPIRY", "15m"))
	if err != nil {
		return nil, fmt.Errorf("invalid ACCESS_TOKEN_EXPIRY: %w", err)
	}
	cfg.AccessTokenExpiry = accessExp

	refreshExp, err := time.ParseDuration(getEnvOrDefault("REFRESH_TOKEN_EXPIRY", "168h"))
	if err != nil {
		return nil, fmt.Errorf("invalid REFRESH_TOKEN_EXPIRY: %w", err)
	}
	cfg.RefreshTokenExpiry = refreshExp

	// Validate required fields
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	return cfg, nil
}

func getEnvOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
