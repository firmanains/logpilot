package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port          string
	LaravelAPIURL string
	ClickUpAPIKey string
	SendGridKey   string
}

func Load() *Config {
	_ = godotenv.Load()
	return &Config{
		Port:          getEnv("DISPATCHER_PORT", "9090"),
		LaravelAPIURL: getEnv("LARAVEL_API_URL", "http://localhost:8000"),
		ClickUpAPIKey: getEnv("CLICKUP_API_KEY", ""),
		SendGridKey:   getEnv("SENDGRID_API_KEY", ""),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
