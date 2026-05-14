package enricher

import (
	"time"

	"github.com/firmanains/logpilot/services/ingestor/internal/domain"
)

type Enricher struct {
	ingestorID string
}

func New(ingestorID string) *Enricher {
	return &Enricher{ingestorID: ingestorID}
}

func (e *Enricher) Enrich(req domain.IngestRequest, projectID string, parsedTimestamp time.Time) *domain.EnrichedLog {
	return &domain.EnrichedLog{
		Level: req.Level,
		Message: req.Message,
		Timestamp: parsedTimestamp,
		Metadata: req.Metadata,
		ProjectID: projectID,
		Service: req.Service,
		IngestedAt: time.Now().UTC(),
		IngestorID: e.ingestorID,
	}
}
