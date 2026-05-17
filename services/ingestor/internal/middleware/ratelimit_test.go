package middleware_test

import (
	"context"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/firmanains/logpilot/services/ingestor/internal/middleware"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func (r *MockRedisClient) Incr(ctx context.Context, key string) *redis.IntCmd {
	cmd := redis.NewIntCmd(ctx)

	if r.err != nil {
		cmd.SetErr(r.err)
	} else {
		cmd.SetVal(int64(r.incrVal))
	}
	return cmd
}

func (r *MockRedisClient) Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd {
	cmd := redis.NewBoolCmd(ctx)

	if r.err != nil {
		cmd.SetErr(r.err)
	}
	return cmd
}

func TestRateLimit(t *testing.T){
	tests := []struct {
		name string
		apiKey string
		wantStatus int
		Redis *MockRedisClient
	}{
		{
			name: "success",
			apiKey: "test-key",
			wantStatus: 200,
			Redis: &MockRedisClient{
				incrVal: 1,
			},
		},
		{
			name: "error: failed to increment",
			apiKey: "test-key",
			wantStatus: 500,
			Redis: &MockRedisClient{
				err: errors.New("error failed to increment"),
			},
		},
		{
			name: "error: rate limit exceeded",
			apiKey: "test-key",
			wantStatus: 429,
			Redis: &MockRedisClient{
				projectID: "test-project",
				incrVal: 10001,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T){
			app := fiber.New()
			app.Use(middleware.RateLimit(tt.Redis))
			app.Post("/test", func(c *fiber.Ctx) error {
				c.Locals(middleware.LocalKeyProjectID, "test-project")
				return c.SendStatus(fiber.StatusOK)
			})

			req:=httptest.NewRequest("POST", "/test", strings.NewReader(""))
			req.Header.Set("Content-Type", "application/json")
			resp, err := app.Test(req)
			if err != nil {
		      t.Fatalf("unexpected error: %v", err)
			}

			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}
