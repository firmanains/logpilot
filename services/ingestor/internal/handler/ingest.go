package handler

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/firmanains/logpilot/services/ingestor/internal/domain"
	"github.com/firmanains/logpilot/services/ingestor/internal/enricher"
	"github.com/firmanains/logpilot/services/ingestor/internal/middleware"
	"github.com/firmanains/logpilot/services/ingestor/internal/producer"
)

// IngestHandler handles POST /v1/ingest.
type IngestHandler struct {
	enricher *enricher.Enricher
	producer *producer.KafkaProducer
	logger   *zap.Logger
}

func NewIngestHandler(e *enricher.Enricher, p *producer.KafkaProducer, log *zap.Logger) *IngestHandler {
	return &IngestHandler{enricher: e, producer: p, logger: log}
}

func (h *IngestHandler) Handle(c *fiber.Ctx) error {
	var req domain.IngestRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid json body",
		})
	}

	if errs := validate(req); len(errs) > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":  "validation failed",
			"fields": errs,
		})
	}

	projectID, _ := c.Locals(middleware.LocalKeyProjectID).(string)
	enriched := h.enricher.Enrich(req, projectID)

	if err := h.producer.Publish(enriched); err != nil {
		h.logger.Error("kafka publish failed", zap.Error(err))
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "failed to process log, please retry",
		})
	}

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{"status": "accepted"})
}

func validate(req domain.IngestRequest) map[string]string {
	errs := map[string]string{}
	if req.Message == "" {
		errs["message"] = "required"
	}
	if req.Service == "" {
		errs["service"] = "required"
	}
	if _, ok := domain.ValidLevels[req.Level]; !ok {
		errs["level"] = fmt.Sprintf("must be one of DEBUG|INFO|WARN|ERROR|FATAL, got %q", req.Level)
	}
	if _, err := time.Parse(time.RFC3339, req.Timestamp); err != nil {
		errs["timestamp"] = "must be ISO 8601 format, e.g. 2026-04-17T05:12:00Z"
	}
	return errs
}
