package domain

import "time"

// LogLevel represents allowed log severity levels.
type LogLevel string

const (
	LevelDebug LogLevel = "DEBUG"
	LevelInfo  LogLevel = "INFO"
	LevelWarn  LogLevel = "WARN"
	LevelError LogLevel = "ERROR"
	LevelFatal LogLevel = "FATAL"
)

// ValidLevels for O(1) lookup during validation.
var ValidLevels = map[LogLevel]struct{}{
	LevelDebug: {}, LevelInfo: {}, LevelWarn: {},
	LevelError: {}, LevelFatal: {},
}

// IngestRequest is the raw payload received from the client.
type IngestRequest struct {
	Level     LogLevel               `json:"level"`
	Message   string                 `json:"message"`
	Service   string                 `json:"service"`
	Timestamp string                 `json:"timestamp"` // ISO 8601 string
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// EnrichedLog is the payload published to Kafka after enrichment.
type EnrichedLog struct {
	Level      LogLevel               `json:"level"`
	Message    string                 `json:"message"`
	Service    string                 `json:"service"`
	Timestamp  time.Time              `json:"timestamp"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	ProjectID  string                 `json:"project_id"`
	IngestedAt time.Time              `json:"ingested_at"`
	IngestorID string                 `json:"ingestor_id"`
}
