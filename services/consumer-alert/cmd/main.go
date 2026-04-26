package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/firmanains/logpilot/services/consumer-alert/internal/config"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	cfg := config.Load()

	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		logger.Fatal("redis connection failed", zap.Error(err))
	}
	logger.Info("redis connected")

	// TODO (Chunk 3.2): Initialize in-memory rule cache
	// TODO (Chunk 3.5): Initialize Kafka alert-events producer
	// TODO (Chunk 3.6): Start Kafka consumer group "alert-evaluators"

	logger.Info("consumer-alert running",
		zap.String("group", cfg.KafkaGroupID),
		zap.String("topic", cfg.KafkaTopicRawLogs),
	)

	_ = rdb

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	logger.Info("shutting down...")
}
