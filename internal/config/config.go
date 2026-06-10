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
	AppPublicHost      string // host[:port] served to Swagger UI, e.g. "api.example.com"
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

	port := getEnv("APP_PORT", "8080")
	cfg := &Config{
		AppEnv:             getEnv("APP_ENV", "development"),
		AppPort:            port,
		AppPublicHost:      getEnv("APP_PUBLIC_HOST", "localhost:"+port),
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
		fmt.Fprintf(os.Stderr, "required env var %q is not set\n", key)
		os.Exit(1)
	}
	return v
}
