// Package config loads runtime configuration from environment variables.
// No secrets are hardcoded here -- fixes the legacy pattern of Telegram bot
// tokens and office coordinates committed directly in PHP source (A6, D-1).
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port string

	DatabaseURL string

	JWTSecret       string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration

	LoginRateLimitPerMinute int // D-4

	StorageDir string // local disk root for check-in photos, D-11

	// FlagJobInterval/FlagJobStaleAfter drive the D-17 "flag no-checkout"
	// background job wired in cmd/api/main.go. A real production scheduler
	// (external cron, etc.) is a Phase 5 decision still open -- this
	// in-process ticker is a dev/demo stand-in, not the final answer.
	FlagJobInterval   time.Duration
	FlagJobStaleAfter time.Duration
}

func Load() (Config, error) {
	cfg := Config{
		Port:                    getEnv("PORT", "8080"),
		DatabaseURL:             os.Getenv("DATABASE_URL"),
		JWTSecret:               os.Getenv("JWT_SECRET"),
		AccessTokenTTL:          getEnvSeconds("ACCESS_TOKEN_TTL_SECONDS", 15*60),
		RefreshTokenTTL:         30 * 24 * time.Hour,
		LoginRateLimitPerMinute: 5,
		StorageDir:              getEnv("STORAGE_DIR", "storage"),
		FlagJobInterval:         getEnvMinutes("FLAG_JOB_INTERVAL_MINUTES", 60),
		FlagJobStaleAfter:       getEnvMinutes("FLAG_JOB_STALE_AFTER_MINUTES", 12*60),
	}

	if cfg.DatabaseURL == "" {
		return cfg, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.JWTSecret == "" {
		return cfg, fmt.Errorf("JWT_SECRET is required")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvMinutes(key string, fallbackMinutes int) time.Duration {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return time.Duration(n) * time.Minute
		}
	}
	return time.Duration(fallbackMinutes) * time.Minute
}

func getEnvSeconds(key string, fallbackSeconds int) time.Duration {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return time.Duration(n) * time.Second
		}
	}
	return time.Duration(fallbackSeconds) * time.Second
}
