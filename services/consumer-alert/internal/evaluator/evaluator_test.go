package evaluator_test

import (
	"testing"

	"github.com/firmanains/logpilot/services/consumer-alert/internal/domain"
	"github.com/firmanains/logpilot/services/consumer-alert/internal/evaluator"
)

func TestEvaluate_MatchesCorrectProjectAndLevel(t *testing.T) {
	rules := []domain.AlertRule{
		{ID: "r1", ProjectID: "proj-a", Condition: domain.RuleCondition{Level: "ERROR"}, IsActive: true, Threshold: 10, WindowSeconds: 300, CooldownSeconds: 600},
		{ID: "r2", ProjectID: "proj-b", Condition: domain.RuleCondition{Level: "ERROR"}, IsActive: true, Threshold: 5, WindowSeconds: 60, CooldownSeconds: 120},
	}
	log := domain.EnrichedLog{ProjectID: "proj-a", Level: "ERROR", Service: "api"}
	matched := evaluator.Evaluate(log, rules)
	if len(matched) != 1 || matched[0].ID != "r1" {
		t.Errorf("expected rule r1 to match, got %v", matched)
	}
}

func TestEvaluate_InactiveRuleSkipped(t *testing.T) {
	rules := []domain.AlertRule{
		{ID: "r1", ProjectID: "proj-a", Condition: domain.RuleCondition{Level: "ERROR"}, IsActive: false},
	}
	log := domain.EnrichedLog{ProjectID: "proj-a", Level: "ERROR"}
	if matched := evaluator.Evaluate(log, rules); len(matched) != 0 {
		t.Error("inactive rule should not match")
	}
}

func TestEvaluate_ServiceFilterApplied(t *testing.T) {
	rules := []domain.AlertRule{
		{ID: "r1", ProjectID: "proj-a", Condition: domain.RuleCondition{Level: "ERROR", Service: "payment-svc"}, IsActive: true},
	}
	logNoMatch := domain.EnrichedLog{ProjectID: "proj-a", Level: "ERROR", Service: "api-gateway"}
	if matched := evaluator.Evaluate(logNoMatch, rules); len(matched) != 0 {
		t.Error("should not match different service")
	}
	logMatch := domain.EnrichedLog{ProjectID: "proj-a", Level: "ERROR", Service: "payment-svc"}
	if matched := evaluator.Evaluate(logMatch, rules); len(matched) != 1 {
		t.Error("should match exact service")
	}
}
