package destinations

import (
	"fmt"
	"time"

	"github.com/firmanains/logpilot/services/alert-dispatcher/internal/domain"
)

// SendGrid sends an alert notification via email.
// TODO (Chunk 3.10): implement using sendgrid-go SDK
func SendGrid(apiKey string, event domain.AlertEvent, cfg domain.NotifConfig) error {
	if len(cfg.EmailRecipients) == 0 {
		return nil
	}
	// Placeholder
	_ = fmt.Sprintf("[LogPilot Alert] %s triggered at %s for project %s. Log: %s",
		event.RuleName, event.TriggeredAt.Format(time.RFC3339), event.ProjectID, event.LogSample)
	return nil
}
