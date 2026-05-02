package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("failed to init logger: %v", err)
	}
	defer logger.Sync()

	app := fiber.New()
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
			"ingestor_id": "test-health",
		})
	})
	logger.Info("Running on port 8080")
	err = app.Listen(":8080")
	if err != nil {
		logger.Fatal("Failed to listen to port 8080: ", zap.Error(err))
	}
}
