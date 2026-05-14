package handler

import (
	"github.com/firmanains/logpilot/services/ingestor/internal/enricher"
	"github.com/firmanains/logpilot/services/ingestor/internal/producer"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

type Handler struct {
	logger   *zap.Logger
	producer *producer.KafkaProducer
	enricher *enricher.Enricher
	validator *validator.Validate
}

func New(logger *zap.Logger, producer *producer.KafkaProducer, enricher *enricher.Enricher, validator *validator.Validate) *Handler {
	return &Handler{
		logger:   logger,
		producer: producer,
		enricher: enricher,
		validator: validator,
	}
}
