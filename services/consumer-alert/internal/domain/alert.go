package domain

import "time"

// AlertRule defines when an alert should fire.
type AlertRule struct {
	ID              string        `json:"id"`
	ProjectID       string        `json:"project_id"`
	Name            string        `json:"name"`
	Condition       RuleCondition `json:"condition"`
	Threshold       int64         `json:"threshold"`
	WindowSeconds   int           `json:"window_seconds"`
	CooldownSeconds int           `json:"cooldown_seconds"`
	IsActive        bool          `json:"is_active"`
}

// RuleCondition specifies which log fields must match.
type RuleCondition struct {
	Level   string `json:"level"`
	Service string `json:"service,omitempty"` // empty = match any service
}

// EnrichedLog mirrors what we consume from Kafka.
type EnrichedLog struct {
	Level     string `json:"level"`
	Message   string `json:"message"`
	Service   string `json:"service"`
	ProjectID string `json:"project_id"`
}

// AlertEvent is published to Kafka when a rule fires.
type AlertEvent struct {
	RuleID      string    `json:"rule_id"`
	ProjectID   string    `json:"project_id"`
	RuleName    string    `json:"rule_name"`
	TriggeredAt time.Time `json:"triggered_at"`
	LogSample   string    `json:"log_sample"`
}
