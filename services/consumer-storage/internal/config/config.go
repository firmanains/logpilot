package config

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

type Config struct {
	Brokers    []string
	Topics     []string
	DBname     string
	DBAddress  string
	DBUsername string
	DBPassword string
	GroupID    string
}

func LoadConfig(logger *zap.Logger) *Config {
	err := godotenv.Load()
	if err != nil {
		logger.Fatal("Error loading .env file", zap.Error(err))
	}

	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		brokers = "localhost:9092"
	}
	brokersArr := strings.Split(brokers, ",")

	topics := os.Getenv("KAFKA_TOPICS")
	if topics == "" {
		topics = "raw-logs"
	}
	topicsArr := strings.Split(topics, ",")

	dbName := os.Getenv("CLICKHOUSE_DB")
	if dbName == "" {
		dbName = "logpilot"
	}

	dbAddress := os.Getenv("CLICKHOUSE_ADDR")
	if dbAddress == "" {
		dbAddress = "localhost:9000"
	}

	dbUsername := os.Getenv("CLICKHOUSE_USERNAME")
	if dbUsername == "" {
		dbUsername = "default"
	}

	dbPassword := os.Getenv("CLICKHOUSE_PASSWORD")

	groupID := os.Getenv("KAFKA_GROUP_ID")
	if groupID == "" {
		groupID = "storage-workers"
	}

	return &Config{
		Brokers:    brokersArr,
		Topics:     topicsArr,
		DBname:     dbName,
		DBAddress:  dbAddress,
		DBUsername: dbUsername,
		DBPassword: dbPassword,
		GroupID:    groupID,
	}

}
