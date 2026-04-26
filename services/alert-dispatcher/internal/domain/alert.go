package domain

import "time"

// AlertEvent received from Alertmanager webhook.
type AlertEvent struct {
	RuleID      string    `json:"rule_id"`
	ProjectID   string    `json:"project_id"`
	RuleName    string    `json:"rule_name"`
	TriggeredAt time.Time `json:"triggered_at"`
	LogSample   string    `json:"log_sample"`
}

// NotifConfig holds notification destinations for a project.
type NotifConfig struct {
	ClickUpListID     string   `json:"clickup_list_id"`
	ClickUpAssigneeID string   `json:"clickup_assignee_id"`
	EmailRecipients   []string `json:"email_recipients"`
	SlackWebhookURL   string   `json:"slack_webhook_url"`
}
