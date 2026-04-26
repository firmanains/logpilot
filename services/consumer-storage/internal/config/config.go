package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	KafkaBrokers    []string
	KafkaTopic      string
	KafkaGroupID    string
	ClickHouseAddr  string
	ClickHouseDB    string
	BatchSize       int
	FlushIntervalMs int
}

func Load() *Config {
	_ = godotenv.Load()
	return &Config{
		KafkaBrokers:    []string{getEnv("KAFKA_BROKERS", "localhost:9092")},
		KafkaTopic:      getEnv("KAFKA_TOPIC_RAW_LOGS", "raw-logs"),
		KafkaGroupID:    "storage-workers",
		ClickHouseAddr:  getEnv("CLICKHOUSE_ADDR", "localhost:9000"),
		ClickHouseDB:    getEnv("CLICKHOUSE_DB", "logpilot"),
		BatchSize:       500,
		FlushIntervalMs: 1000,
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
