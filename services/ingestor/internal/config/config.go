package config

import (
	"os"

	"github.com/joho/godotenv"
)

// Config holds all runtime configuration for the ingestor.
type Config struct {
	Port         string
	RedisAddr    string
	KafkaBrokers []string
	KafkaTopic   string
	RateLimitMax int64
	IngestorID   string
}

// Load reads configuration from environment variables.
func Load() *Config {
	_ = godotenv.Load()
	return &Config{
		Port:         getEnv("INGESTOR_PORT", "8080"),
		RedisAddr:    getEnv("REDIS_ADDR", "localhost:6379"),
		KafkaBrokers: []string{getEnv("KAFKA_BROKERS", "localhost:9092")},
		KafkaTopic:   getEnv("KAFKA_TOPIC_RAW_LOGS", "raw-logs"),
		RateLimitMax: 10000,
		IngestorID:   getHostname(),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getHostname() string {
	h, _ := os.Hostname()
	return h
}
