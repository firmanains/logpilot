package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/firmanains/logpilot/services/ingestor/internal/config"
	"github.com/firmanains/logpilot/services/ingestor/internal/enricher"
	"github.com/firmanains/logpilot/services/ingestor/internal/handler"
	"github.com/firmanains/logpilot/services/ingestor/internal/middleware"
	"github.com/firmanains/logpilot/services/ingestor/internal/producer"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	cfg := config.Load()
	logger.Info("ingestor starting", zap.String("port", cfg.Port))

	// ── Redis ────────────────────────────────────────────────
	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		logger.Fatal("redis connection failed", zap.Error(err))
	}
	logger.Info("redis connected", zap.String("addr", cfg.RedisAddr))

	// ── Kafka Producer ───────────────────────────────────────
	kafkaProducer, err := producer.New(cfg.KafkaBrokers, cfg.KafkaTopic, logger)
	if err != nil {
		logger.Fatal("kafka producer init failed", zap.Error(err))
	}
	defer kafkaProducer.Close()
	logger.Info("kafka producer connected")

	// ── Wire Dependencies ────────────────────────────────────
	logEnricher := enricher.New(cfg.IngestorID)
	ingestHandler := handler.NewIngestHandler(logEnricher, kafkaProducer, logger)

	// ── HTTP Server ──────────────────────────────────────────
	app := fiber.New(fiber.Config{
		AppName:      "LogPilot Ingestor",
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	})

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "ingestor_id": cfg.IngestorID})
	})

	v1 := app.Group("/v1")
	v1.Post("/ingest",
		middleware.Auth(rdb),
		middleware.RateLimit(rdb, cfg.RateLimitMax),
		ingestHandler.Handle,
	)

	// ── Graceful Shutdown ────────────────────────────────────
	go func() {
		if err := app.Listen(":" + cfg.Port); err != nil {
			logger.Fatal("server error", zap.Error(err))
		}
	}()

	logger.Info("ingestor ready", zap.String("port", cfg.Port))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down...")
	app.Shutdown()
}
