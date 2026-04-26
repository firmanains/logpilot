package evaluator

import "github.com/firmanains/logpilot/services/consumer-alert/internal/domain"

// Evaluate returns all active rules that match the given log.
func Evaluate(log domain.EnrichedLog, rules []domain.AlertRule) []domain.AlertRule {
	var matched []domain.AlertRule
	for _, rule := range rules {
		if !rule.IsActive {
			continue
		}
		if rule.ProjectID != log.ProjectID {
			continue
		}
		if rule.Condition.Level != log.Level {
			continue
		}
		if rule.Condition.Service != "" && rule.Condition.Service != log.Service {
			continue
		}
		matched = append(matched, rule)
	}
	return matched
}
