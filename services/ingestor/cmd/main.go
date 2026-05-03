package main

import (
	"context"
	"log"

	"github.com/firmanains/logpilot/services/ingestor/internal/config"
	"github.com/firmanains/logpilot/services/ingestor/internal/middleware"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("failed to init logger: %v", err)
	}
	defer logger.Sync()

	cfg := config.LoadConfig(logger)

	// new redis connection
	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisAddress})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		logger.Fatal("failed to connect to redis", zap.Error(err))
	}
	logger.Info("redis connected", zap.String("addr", cfg.RedisAddress))

	app := fiber.New()
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
			"ingestor_id": "test-health",
		})
	})

	v1 := app.Group("/v1")
	v1.Post("/ingest",
		middleware.Authentication(rdb),
		func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusOK).JSON(fiber.Map{
				"message": "test aja placeholder",
			})
		},
	)

	logger.Info("Running on port 8080")
	err = app.Listen(":8080")
	if err != nil {
		logger.Fatal("Failed to listen to port 8080: ", zap.Error(err))
	}
}
