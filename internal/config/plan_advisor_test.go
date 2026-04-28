package config

import "testing"

func TestPlanAdvisorDefaultsAndNormalization(t *testing.T) {
	cfg := &Config{}
	applyDefaults(cfg)
	if cfg.Workflow.PlanAdvisorMode != PlanAdvisorMissingInfo || cfg.Workflow.PlanQuestionLimit != DefaultPlanQuestionCap {
		t.Fatalf("unexpected advisor defaults: %+v", cfg.Workflow)
	}
	if NormalizePlanAdvisorMode("off") != PlanAdvisorOff || NormalizePlanAdvisorMode("ALWAYS") != PlanAdvisorAlways {
		t.Fatal("expected advisor modes to normalize")
	}
	if NormalizePlanQuestionLimit(9) != DefaultPlanQuestionCap || NormalizePlanQuestionLimit(2) != 2 {
		t.Fatal("expected question limit normalization")
	}
}
