package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

// RateLimit enforces per-project request limits.
// Window: 60 seconds. Redis key: rate:{project_id}
func RateLimit(rdb *redis.Client, maxReqPerMinute int64) fiber.Handler {
	return func(c *fiber.Ctx) error {
		projectID, _ := c.Locals(LocalKeyProjectID).(string)
		key := fmt.Sprintf("rate:%s", projectID)
		ctx := context.Background()

		count, err := rdb.Incr(ctx, key).Result()
		if err != nil {
			return c.Next() // fail open if Redis is unavailable
		}
		if count == 1 {
			rdb.Expire(ctx, key, 60*time.Second)
		}
		if count > maxReqPerMinute {
			c.Set("Retry-After", "60")
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":       "rate limit exceeded",
				"retry_after": 60,
			})
		}
		return c.Next()
	}
}
