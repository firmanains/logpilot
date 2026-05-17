package middleware_test

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/firmanains/logpilot/services/ingestor/internal/middleware"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

type MockRedisClient struct {
	projectID string
	err error
	incrVal int
}

func (r *MockRedisClient) Get(ctx context.Context, key string) *redis.StringCmd {
	cmd := redis.NewStringCmd(ctx)

	if r.err != nil {
		cmd.SetErr(r.err)
	} else {
		cmd.SetVal(r.projectID)
	}
	return cmd
}

func TestAuthentication(t *testing.T){
	tests := []struct {
		name string
		apiKey string
		wantStatus int
		Redis *MockRedisClient
	}{
		{
			name: "successfully authenticated",
			apiKey: "test-key",
			wantStatus: 200,
			Redis: &MockRedisClient{
				projectID: "test-project",
				err: nil,
			},
		},
		{
			name: "apikey not found",
			apiKey: "",
			wantStatus: 401,
			Redis: &MockRedisClient{},
		},
		{
			name: "invalid key",
			apiKey: "invalid*(&*key",
			wantStatus: 401,
			Redis: &MockRedisClient{
				projectID: "test-project",
				err: redis.Nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T){
			app := fiber.New()
			app.Use(middleware.Authentication(tt.Redis))
			app.Post("/test", func(c *fiber.Ctx) error {
				return c.SendStatus(fiber.StatusOK)
			})

			req:=httptest.NewRequest("POST", "/test", strings.NewReader(""))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-API-KEY", tt.apiKey)
			resp, err := app.Test(req)
			if err != nil {
		      t.Fatalf("unexpected error: %v", err)
			}

			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}
