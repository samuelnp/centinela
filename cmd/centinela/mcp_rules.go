package main

import (
	"github.com/samuelnp/centinela/internal/config"
	mcpgov "github.com/samuelnp/centinela/internal/mcp"
	"github.com/samuelnp/centinela/internal/workflow"
)

// mcpRules reads the governing rule surface from config + the active archetype.
// Best-effort: a missing/invalid config yields the defaults rather than failing
// the read_rules tool.
func mcpRules() mcpgov.RulesOutput {
	out := mcpgov.RulesOutput{MaxFileLines: 100}
	if cfg, err := config.Load(); err == nil {
		out.Profile = cfg.Workflow.EnforcementProfile
		out.Locales = cfg.I18n.Locales
		out.Gates = enabledGateNames(cfg)
	}
	arch, _ := workflow.DisplayArchetype(activeWorkflow(mustGetwd()))
	out.Archetype = arch
	return out
}

func enabledGateNames(cfg *config.Config) []string {
	var names []string
	add := func(on bool, name string) {
		if on {
			names = append(names, name)
		}
	}
	add(cfg.Gates.FileSizeEnabled, "file_size")
	add(cfg.Gates.Build.Enabled, "build")
	add(cfg.Gates.ImportGraph.Enabled, "import_graph")
	add(cfg.Gates.Security.Enabled, "security")
	add(cfg.Gates.SpecTraceability.Enabled, "spec_traceability")
	add(cfg.Gates.RoadmapDrift.Enabled, "roadmap_drift")
	add(cfg.Gates.AuditBaseline.Enabled, "audit_baseline")
	return names
}
