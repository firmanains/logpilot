package middleware

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

const LocalKeyProjectID = "projectID"

// Auth validates X-API-Key header against Redis.
// Redis key pattern: api_key:{plain_key} → project_id
func Auth(rdb *redis.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		key := c.Get("X-API-Key")
		if key == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "missing X-API-Key header",
			})
		}

		projectID, err := rdb.Get(context.Background(), fmt.Sprintf("api_key:%s", key)).Result()
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid api key",
			})
		}

		c.Locals(LocalKeyProjectID, projectID)
		return c.Next()
	}
}
