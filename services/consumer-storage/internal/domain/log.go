package domain

import "time"

// EnrichedLog mirrors the ingestor output — what we read from Kafka.
type EnrichedLog struct {
	Level      string                 `json:"level"`
	Message    string                 `json:"message"`
	Service    string                 `json:"service"`
	Timestamp  time.Time              `json:"timestamp"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	ProjectID  string                 `json:"project_id"`
	IngestedAt time.Time              `json:"ingested_at"`
	IngestorID string                 `json:"ingestor_id"`
}
