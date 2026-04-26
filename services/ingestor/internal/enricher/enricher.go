package enricher

import (
	"time"

	"github.com/firmanains/logpilot/services/ingestor/internal/domain"
)

// Enricher adds server-side fields to incoming log requests.
type Enricher struct {
	ingestorID string
}

func New(ingestorID string) *Enricher {
	return &Enricher{ingestorID: ingestorID}
}

// Enrich creates an EnrichedLog from a raw request.
func (e *Enricher) Enrich(req domain.IngestRequest, projectID string) domain.EnrichedLog {
	ts, _ := time.Parse(time.RFC3339, req.Timestamp)
	return domain.EnrichedLog{
		Level:      req.Level,
		Message:    req.Message,
		Service:    req.Service,
		Timestamp:  ts,
		Metadata:   req.Metadata,
		ProjectID:  projectID,
		IngestedAt: time.Now().UTC(),
		IngestorID: e.ingestorID,
	}
}
