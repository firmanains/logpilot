package main

import (
	"context"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/IBM/sarama"
	"github.com/firmanains/logpilot/services/consumer-storage/internal/config"
	"github.com/firmanains/logpilot/services/consumer-storage/internal/consumer"
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
		logger.Fatal("failed to connect to DB", zap.Error(err))
	}
	logger.Info("successfully connected to DB")
	defer conn.Close()

	cfgSarama := sarama.NewConfig()
	cfgSarama.Consumer.Offsets.Initial = sarama.OffsetOldest

	group, err := sarama.NewConsumerGroup(cfg.Brokers, cfg.GroupID, cfgSarama)
	if err != nil {
		logger.Fatal("failed to create consumer group", zap.Error(err))
	}
	defer group.Close()

	handler := consumer.NewConsumer()

	for {
		group.Consume(context.Background(), cfg.Topics, handler)
	}

}
