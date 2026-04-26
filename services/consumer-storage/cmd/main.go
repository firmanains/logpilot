package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/IBM/sarama"
	"go.uber.org/zap"

	"github.com/firmanains/logpilot/services/consumer-storage/internal/config"
	"github.com/firmanains/logpilot/services/consumer-storage/internal/consumer"
	"github.com/firmanains/logpilot/services/consumer-storage/internal/storage"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	cfg := config.Load()

	store, err := storage.New(cfg.ClickHouseAddr, cfg.ClickHouseDB, logger)
	if err != nil {
		logger.Fatal("clickhouse connection failed", zap.Error(err))
	}
	defer store.Close()

	if err := store.EnsureSchema(context.Background()); err != nil {
		logger.Fatal("schema creation failed", zap.Error(err))
	}
	logger.Info("clickhouse ready")

	kafkaCfg := sarama.NewConfig()
	kafkaCfg.Consumer.Offsets.Initial = sarama.OffsetOldest
	kafkaCfg.Consumer.Offsets.AutoCommit.Enable = false

	cg, err := sarama.NewConsumerGroup(cfg.KafkaBrokers, cfg.KafkaGroupID, kafkaCfg)
	if err != nil {
		logger.Fatal("kafka consumer group init failed", zap.Error(err))
	}
	defer cg.Close()

	handler := consumer.NewHandler(store, logger, cfg.BatchSize, cfg.FlushIntervalMs)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		for {
			if err := cg.Consume(ctx, []string{cfg.KafkaTopic}, handler); err != nil {
				logger.Error("consumer group error", zap.Error(err))
			}
			if ctx.Err() != nil {
				return
			}
		}
	}()

	logger.Info("consumer-storage running",
		zap.String("group", cfg.KafkaGroupID),
		zap.String("topic", cfg.KafkaTopic),
	)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down...")
	cancel()
}
