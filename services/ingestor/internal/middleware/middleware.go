package middleware

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)
const LocalKeyProjectID = "projectID"
type RedisClient interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	Incr(ctx context.Context, key string) *redis.IntCmd
	Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd
}
