package handler

import (
	"time"

	"github.com/firmanains/logpilot/services/ingestor/internal/domain"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

type Handler struct {
	logger   *zap.Logger
	producer Producer
	enricher Enricher
	validator *validator.Validate
}

type Producer interface {
	Publish(log *domain.EnrichedLog) error
}

type Enricher interface {
	Enrich(req domain.IngestRequest, projectID string, parsedTimestamp time.Time) *domain.EnrichedLog
}

func New(logger *zap.Logger, producer Producer, enricher Enricher, validator *validator.Validate) *Handler {
	return &Handler{
		logger:   logger,
		producer: producer,
		enricher: enricher,
		validator: validator,
	}
}
