package config

// applyDefaults normalizes every config section in place, filling unset fields
// with their defaults so downstream consumers never see a zero value that would
// otherwise change behavior (e.g. a blank diff mode or an unconfigured gate).
func applyDefaults(cfg *Config) {
	if cfg.Gates.FileSizeEnabled == false && cfg.Gates.I18nEnabled == false {
		cfg.Gates.FileSizeEnabled = true
	}
	cfg.Workflow.StepConfirmationMode = NormalizeStepConfirmationMode(cfg.Workflow.StepConfirmationMode)
	cfg.Workflow.PlanAdvisorMode = NormalizePlanAdvisorMode(cfg.Workflow.PlanAdvisorMode)
	cfg.Workflow.EnforcementProfile = NormalizeEnforcementProfile(cfg.Workflow.EnforcementProfile)
	cfg.Workflow.PlanQuestionLimit = NormalizePlanQuestionLimit(cfg.Workflow.PlanQuestionLimit)
	cfg.Workflow.PlanAdvisorFailureTopN = NormalizePlanAdvisorFailureTopN(cfg.Workflow.PlanAdvisorFailureTopN)
	cfg.Validate.DiffMode = NormalizeDiffMode(cfg.Validate.DiffMode)
	cfg.Validate.DiffBase = NormalizeDiffBase(cfg.Validate.DiffBase)
	applyMemoryDefaults(cfg)
	if cfg.Verify.TimeoutSeconds <= 0 {
		cfg.Verify.TimeoutSeconds = 60
	}
	if cfg.Verify.CoverageTolerance <= 0 {
		cfg.Verify.CoverageTolerance = 0.001
	}
	cfg.Gates.Build = NormalizeBuildGate(cfg.Gates.Build)
	cfg.Gates.ImportGraph = NormalizeImportGraph(cfg.Gates.ImportGraph)
	cfg.Gates.Security = NormalizeSecurityGate(cfg.Gates.Security)
	cfg.Gates.SpecTraceability = NormalizeSpecTraceability(cfg.Gates.SpecTraceability)
	cfg.Gates.RoadmapDrift = NormalizeRoadmapDrift(cfg.Gates.RoadmapDrift)
	cfg.Gates.AuditBaseline = NormalizeAuditBaseline(cfg.Gates.AuditBaseline)
	cfg.Gates.CustomGates = NormalizeCustomGates(cfg.Gates.CustomGates)
	cfg.Precommit = NormalizePrecommit(cfg.Precommit)
	cfg.PrGate = NormalizePrGate(cfg.PrGate)
	cfg.Cost = NormalizeCost(cfg.Cost)
}
