package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv             string
	AppPort            string
	DatabaseURL        string
	JWTSecret          string
	JWTAccessTokenTTL  time.Duration
	CORSAllowedOrigins []string
}

func Load() (*Config, error) {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}
	if env != "production" {
		_ = godotenv.Load()
	}

	ttl, err := time.ParseDuration(getEnv("JWT_ACCESS_TOKEN_TTL", "24h"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_ACCESS_TOKEN_TTL: %w", err)
	}

	origins := strings.Split(getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000"), ",")

	cfg := &Config{
		AppEnv:             getEnv("APP_ENV", "development"),
		AppPort:            getEnv("APP_PORT", "8080"),
		DatabaseURL:        mustGetEnv("DATABASE_URL"),
		JWTSecret:          mustGetEnv("JWT_SECRET"),
		JWTAccessTokenTTL:  ttl,
		CORSAllowedOrigins: origins,
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func mustGetEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("required env var %q is not set", key))
	}
	return v
}
