package config

import (
	"os"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

type Config struct {
	RedisAddress string
	KafkaBrokers string
	KafkaTopic string
	Port string
}

func LoadConfig(logger *zap.Logger) *Config {
	err := godotenv.Load()
	if err != nil {
		logger.Info("failed to load config", zap.Error(err))
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	if kafkaBrokers == "" {
		kafkaBrokers = "localhost:9092"
	}

	kafkaTopic := os.Getenv("KAFKA_TOPIC_RAW_LOG")
	if kafkaTopic == "" {
		kafkaTopic = "raw-logs"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = ":8080"
	}


	return &Config{
		RedisAddress: redisAddr,
		KafkaBrokers: kafkaBrokers,
		KafkaTopic: kafkaTopic,
		Port: port,
	}
}
