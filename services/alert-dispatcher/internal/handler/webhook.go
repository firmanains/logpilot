package handler

import (
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/firmanains/logpilot/services/alert-dispatcher/internal/destinations"
	"github.com/firmanains/logpilot/services/alert-dispatcher/internal/domain"
)

// WebhookHandler receives Alertmanager webhooks and fans out notifications.
type WebhookHandler struct {
	clickupKey  string
	sendgridKey string
	logger      *zap.Logger
}

func NewWebhookHandler(clickupKey, sendgridKey string, logger *zap.Logger) *WebhookHandler {
	return &WebhookHandler{clickupKey: clickupKey, sendgridKey: sendgridKey, logger: logger}
}

func (h *WebhookHandler) Handle(c *fiber.Ctx) error {
	var event domain.AlertEvent
	if err := c.BodyParser(&event); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid payload"})
	}

	// TODO (Chunk 4.10): fetch real NotifConfig from Laravel API by event.ProjectID
	cfg := domain.NotifConfig{}

	var wg sync.WaitGroup
	dispatch := func(name string, fn func() error) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			done := make(chan error, 1)
			go func() { done <- fn() }()
			select {
			case err := <-done:
				if err != nil {
					h.logger.Error("dispatch failed", zap.String("dest", name), zap.Error(err))
				} else {
					h.logger.Info("dispatched", zap.String("dest", name))
				}
			case <-time.After(10 * time.Second):
				h.logger.Warn("dispatch timeout", zap.String("dest", name))
			}
		}()
	}

	dispatch("clickup", func() error { return destinations.ClickUp(h.clickupKey, event, cfg) })
	dispatch("sendgrid", func() error { return destinations.SendGrid(h.sendgridKey, event, cfg) })
	wg.Wait()

	return c.JSON(fiber.Map{"status": "dispatched"})
}
