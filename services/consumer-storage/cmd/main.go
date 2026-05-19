package main

import (
	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/firmanains/logpilot/services/consumer-storage/internal/config"
	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}

	cfg := config.LoadConfig(logger)

	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{cfg.DBAddress},
		Auth: clickhouse.Auth{
			Database: cfg.DBname,
			Username: cfg.DBUsername,
			Password: cfg.DBPassword,
		},
	})
	if err != nil {
		logger.Error("failed to connect to DB", zap.Error(err))
	}
	logger.Info("successfully connected to DB")
	defer conn.Close()

}
