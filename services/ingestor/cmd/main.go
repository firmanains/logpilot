package main

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/firmanains/logpilot/services/ingestor/internal/config"
	"github.com/firmanains/logpilot/services/ingestor/internal/enricher"
	"github.com/firmanains/logpilot/services/ingestor/internal/handler"
	"github.com/firmanains/logpilot/services/ingestor/internal/middleware"
	"github.com/firmanains/logpilot/services/ingestor/internal/producer"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func main() {
	// init log
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("failed to init logger: %v", err)
	}
	defer logger.Sync()

	// load config
	cfg := config.LoadConfig(logger)

	// new redis connection
	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisAddress})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		logger.Fatal("failed to connect to redis", zap.Error(err))
	}
	logger.Info("redis connected", zap.String("addr", cfg.RedisAddress))

	// ingestor id
	ingestorID, err := os.Hostname()
	if err != nil {
		logger.Error("failed to get ingestor id", zap.Error(err))
		ingestorID = "unknown"
	}
	// enricher
	e := enricher.New(ingestorID)

	// producer
	p, err := producer.NewKafkaProducer(cfg.KafkaTopic, logger, strings.Split(cfg.KafkaBrokers, ","))
	if err != nil {
		logger.Fatal("failed to init kafka producer", zap.Error(err))
	}
	logger.Info("kafka has been initialized", zap.String("kafka", cfg.KafkaBrokers))

	// validator
	v := validator.New()
	// handler
	h := handler.New(logger, p, e, v)

	app := fiber.New()
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":      "ok",
			"ingestor_id": ingestorID,
		})
	})

	v1 := app.Group("/v1")
	v1.Post("/ingest",
		middleware.Authentication(rdb),
		middleware.RateLimit(rdb),
		h.IngestHandle,
	)

	logger.Info("Running on port:", zap.String("PORT", cfg.Port))
	err = app.Listen(cfg.Port)
	if err != nil {
		logger.Fatal("Failed to listen to port", zap.Error(err))
	}
}
