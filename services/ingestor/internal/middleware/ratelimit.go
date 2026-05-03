package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

func RateLimit(rdb *redis.Client) fiber.Handler {
	return func (c *fiber.Ctx) error {
		projectID, _ := c.Locals(LocalKeyProjectID).(string)
		key := fmt.Sprintf("rate:%s", projectID)
		ctx := context.Background()

		val, err := rdb.Incr(ctx, key).Result()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		if val == 1 {
			rdb.Expire(ctx, key, time.Second * 60)
		}

		if val > 10000 {
			c.Set("Retry-After", "60")
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "rate limit exceeded",
			})
		} else {
			return c.Next()
		}
	}
}
