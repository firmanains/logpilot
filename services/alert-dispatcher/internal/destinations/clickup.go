package destinations

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/firmanains/logpilot/services/alert-dispatcher/internal/domain"
)

const clickupBaseURL = "https://api.clickup.com/api/v2"

// ClickUp creates a task in ClickUp for the given alert event.
func ClickUp(apiKey string, event domain.AlertEvent, cfg domain.NotifConfig) error {
	if cfg.ClickUpListID == "" {
		return nil // not configured, skip
	}

	payload := map[string]interface{}{
		"name":     fmt.Sprintf("[ALERT] %s — %s", event.RuleName, event.ProjectID),
		"priority": 1, // Urgent
		"description": fmt.Sprintf(
			"Alert triggered at %s\n\nProject: %s\nRule: %s\n\nLog sample:\n%s",
			event.TriggeredAt.Format(time.RFC3339),
			event.ProjectID, event.RuleName, event.LogSample,
		),
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST",
		fmt.Sprintf("%s/list/%s/task", clickupBaseURL, cfg.ClickUpListID),
		bytes.NewReader(body),
	)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("clickup returned status %d", resp.StatusCode)
	}
	return nil
}
