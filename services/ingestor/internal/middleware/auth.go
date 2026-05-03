package middleware

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

const LocalKeyProjectID = "projectID"

func Authentication(rdb *redis.Client) fiber.Handler {
	return func (c *fiber.Ctx) error {
		apiKey := c.Get("X-API-KEY")
		if apiKey == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "X-API-KEY not found",
			})
		}

		projectKey, err := rdb.Get(context.Background(), fmt.Sprintf("api_key:%s", apiKey)).Result()
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "invalid api key",
			})
		}

		c.Locals(LocalKeyProjectID,  projectKey)
		return c.Next()
	}
}
