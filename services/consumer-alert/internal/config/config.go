package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	KafkaBrokers       []string
	KafkaTopicRawLogs  string
	KafkaTopicAlerts   string
	KafkaGroupID       string
	RedisAddr          string
	LaravelAPIURL      string
}

func Load() *Config {
	_ = godotenv.Load()
	return &Config{
		KafkaBrokers:      []string{getEnv("KAFKA_BROKERS", "localhost:9092")},
		KafkaTopicRawLogs: getEnv("KAFKA_TOPIC_RAW_LOGS", "raw-logs"),
		KafkaTopicAlerts:  getEnv("KAFKA_TOPIC_ALERT_EVENTS", "alert-events"),
		KafkaGroupID:      "alert-evaluators",
		RedisAddr:         getEnv("REDIS_ADDR", "localhost:6379"),
		LaravelAPIURL:     getEnv("LARAVEL_API_URL", "http://localhost:8000"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
