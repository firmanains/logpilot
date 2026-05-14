package domain

import "time"

type LogLevel string

const (
	LevelError LogLevel = "ERROR"
	LevelDebug LogLevel = "DEBUG"
	LevelWarn  LogLevel = "WARN"
	LevelInfo  LogLevel = "INFO"
	LevelFatal LogLevel = "FATAL"
)

type IngestRequest struct {
	Level     LogLevel       `json:"level" validate:"required,oneof=ERROR DEBUG WARN INFO FATAL"`
	Message   string         `json:"message" validate:"required"`
	Service   string         `json:"service" validate:"required"`
	Timestamp string     `json:"timestamp" validate:"required,datetime=2006-01-02T15:04:05Z07:00"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

type EnrichedLog struct {
	Level      LogLevel       `json:"level"`
	Message    string         `json:"message"`
	Timestamp  time.Time      `json:"timestamp"`
	Metadata   map[string]any `json:"metadata"`
	Service    string         `json:"service"`
	ProjectID  string         `json:"project_id"`
	IngestedAt time.Time      `json:"ingested_at"`
	IngestorID string         `json:"ingestor_id"`
}
