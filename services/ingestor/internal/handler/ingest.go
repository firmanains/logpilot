package handler

import (
	"fmt"
	"time"

	"github.com/firmanains/logpilot/services/ingestor/internal/domain"
	"github.com/firmanains/logpilot/services/ingestor/internal/middleware"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func (h *Handler) IngestHandle(c *fiber.Ctx) error {
	var req domain.IngestRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Error("failed to parse Body", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "failed to parse json",
		})
	}


	if err := h.validator.Struct(req); err != nil {
		h.logger.Error("failed to validate payload", zap.Error(err))
		v := err.(validator.ValidationErrors)
		sl := make([]map[string]string, 0, len(v))
		for _, val := range v {
			var field string
			switch val.Tag(){
				case "required":
					field = fmt.Sprintf("%s is required", val.Field())
				case "oneof":
					field = fmt.Sprintf("%s value must be one of ERROR, DEBUG, WARN, INFO, FATAL", val.Field())
				default:
					field = fmt.Sprintf("%s is invalid", val.Field())
			}
			sl = append(sl, map[string]string{
				"field": val.Field(),
				"message": field,
			})
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"errors": sl,
		})
	}
	parsedTime, err := time.Parse("2006-01-02T15:04:05Z07:00", req.Timestamp)

	if err != nil {
		h.logger.Error("failed to parse timestamp", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "failed to parse timestamp",
		})
	}
	projectId := c.Locals(middleware.LocalKeyProjectID).(string)

	enrichedLog := h.enricher.Enrich(req, projectId, parsedTime)
	if err := h.producer.Publish(enrichedLog); err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"message": "failed to publish",
		})
	}

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"message": "log successfully processed",
	})


}
