package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

var (
	ingestorURL string
	apiKey      string
)

func main() {
	_ = godotenv.Load()
	ingestorURL = getEnv("INGESTOR_URL", "http://localhost:8080")
	apiKey = getEnv("API_KEY", "logpilot_changeme")

	app := fiber.New(fiber.Config{AppName: "LogPilot Dummy E-commerce"})

	app.Get("/health", func(c *fiber.Ctx) error {
		sendLog("DEBUG", "health check", "dummy-app", nil)
		return c.JSON(fiber.Map{"status": "ok"})
	})

	app.Get("/products", func(c *fiber.Ctx) error {
		sendLog("INFO", "fetch products success", "dummy-app", nil)
		return c.JSON(fiber.Map{"products": []string{"item-1", "item-2", "item-3"}})
	})

	app.Get("/products/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		if rand.Float32() < 0.2 {
			sendLog("ERROR", fmt.Sprintf("product not found: %s", id), "dummy-app", nil)
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "product not found"})
		}
		sendLog("INFO", fmt.Sprintf("product fetched: %s", id), "dummy-app", nil)
		return c.JSON(fiber.Map{"id": id, "name": "Sample Product", "price": 99.99})
	})

	app.Post("/checkout", func(c *fiber.Ctx) error {
		if rand.Float32() < 0.3 {
			sendLog("ERROR", "payment gateway timeout", "dummy-app", map[string]interface{}{
				"trace_id": fmt.Sprintf("trace-%d", rand.Intn(99999)),
			})
			return c.Status(fiber.StatusGatewayTimeout).JSON(fiber.Map{"error": "payment gateway timeout"})
		}
		orderID := rand.Intn(99999)
		sendLog("INFO", fmt.Sprintf("checkout success, order %d", orderID), "dummy-app", nil)
		return c.JSON(fiber.Map{"status": "paid", "order_id": orderID})
	})

	port := getEnv("APP_PORT", "8081")
	app.Listen(":" + port)
}

func sendLog(level, message, service string, metadata map[string]interface{}) {
	payload := map[string]interface{}{
		"level":     level,
		"message":   message,
		"service":   service,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}
	if metadata != nil {
		payload["metadata"] = metadata
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", ingestorURL+"/v1/ingest", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", apiKey)
	(&http.Client{Timeout: 3 * time.Second}).Do(req)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
