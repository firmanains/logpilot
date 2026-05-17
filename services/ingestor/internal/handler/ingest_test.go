package handler_test

import (
	"errors"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/firmanains/logpilot/services/ingestor/internal/domain"
	"github.com/firmanains/logpilot/services/ingestor/internal/handler"
	"github.com/firmanains/logpilot/services/ingestor/internal/middleware"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type MockEnricher struct {
	enrichedLog *domain.EnrichedLog
}

func (e *MockEnricher) Enrich(req domain.IngestRequest, projectID string, parsedTimestamp time.Time) *domain.EnrichedLog {
	return e.enrichedLog
}

type MockProducer struct {
	publishErr error
}

func (p *MockProducer) Publish(log *domain.EnrichedLog) error {
	return p.publishErr
}


func TestIngestHandler(t *testing.T) {
	tests := []struct {
		name string
		body string
		wantStatus int
		producer *MockProducer
	}{
		{
			name: "valid payload",
			body: `{"level":"INFO","message":"test error","service":"test-svc","timestamp":"2026-05-14T05:12:00Z"}`,
			wantStatus: 202,
			producer: &MockProducer{},
		},
		{
			name: "invalid payload",
			body: `{"level":"INVALID","message":"test supposed to error","service":"test-svc","timestamp":"2026-05-14T05:12:00Z"}`,
			wantStatus: 400,
			producer: &MockProducer{},
		},
		{
			name: "missing field LEVEL",
			body: `"message":"test supposed to error","service":"test-svc","timestamp":"2026-05-14T05:12:00Z"}`,
			wantStatus: 400,
			producer: &MockProducer{},
		},
		{
			name: "valid payload",
			body: `{"level":"INFO","message":"test error","service":"test-svc","timestamp":"2026-05-14T05:12:00Z"}`,
			wantStatus: 503,
			producer: &MockProducer{publishErr: errors.New("error kafka is down")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func (t *testing.T) {
			app := fiber.New()
			app.Use(func(c *fiber.Ctx) error {
				c.Locals(middleware.LocalKeyProjectID, "test-project")
				return c.Next()
			})
			v := validator.New()
			h := handler.New(zap.NewNop(), tt.producer, &MockEnricher{}, v)
			app.Post("/v1/ingest", h.IngestHandle)

			req := httptest.NewRequest("POST", "/v1/ingest", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			resp, err := app.Test(req)
			if err != nil {
		      t.Fatalf("unexpected error: %v", err)
			}

			assert.Equal(t, tt.wantStatus, resp.StatusCode)

		})
	}

}
