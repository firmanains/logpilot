package main

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/firmanains/logpilot/services/alert-dispatcher/internal/config"
	"github.com/firmanains/logpilot/services/alert-dispatcher/internal/handler"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	cfg := config.Load()
	webhookHandler := handler.NewWebhookHandler(cfg.ClickUpAPIKey, cfg.SendGridKey, logger)

	app := fiber.New(fiber.Config{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 15 * time.Second,
	})

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})
	app.Post("/webhook", webhookHandler.Handle)

	logger.Info("alert-dispatcher ready", zap.String("port", cfg.Port))
	app.Listen(":" + cfg.Port)
}
