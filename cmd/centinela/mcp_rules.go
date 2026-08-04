package main

import (
	"github.com/samuelnp/centinela/internal/config"
	mcpgov "github.com/samuelnp/centinela/internal/mcp"
	"github.com/samuelnp/centinela/internal/workflow"
)

// mcpRules reads the governing rule surface from config + the active archetype.
// Best-effort: a missing/invalid config yields the defaults rather than failing
// the read_rules tool.
//
// read_rules is what an agent consults to learn how it is governed, so profile
// and archetype must describe the SAME thing. Both are workflow-scoped when a
// workflow is active and project-scoped only when none is: reporting a
// project-scope profile beside a workflow-scope archetype mixed two scopes in one
// payload and disagreed with status and verdict on every legacy workflow, whose
// pin keeps it strict. Reading cfg.Workflow.EnforcementProfile was wrong the
// other way — it reported applyDefaults' normalized "strict" for any project that
// merely omits the knob. LoadForProfile fails closed, so an unreadable config
// reports strict here exactly as it does everywhere else.
func mcpRules() mcpgov.RulesOutput {
	out := mcpgov.RulesOutput{MaxFileLines: 100}
	cfg, err := config.LoadForProfile()
	wf := activeWorkflow(mustGetwd())
	out.Profile = rulesProfile(wf, cfg)
	if err == nil {
		out.Locales = cfg.I18n.Locales
		out.Gates = enabledGateNames(cfg)
	}
	arch, _ := workflow.DisplayArchetype(wf)
	out.Archetype = arch
	return out
}

// rulesProfile resolves the profile read_rules reports: the one governing the
// active workflow, or the project default when there is no workflow to govern.
// The scope must match the archetype field beside it.
func rulesProfile(wf *workflow.Workflow, cfg *config.Config) string {
	if wf == nil {
		return config.ProjectDefaultProfile(cfg)
	}
	return workflow.EffectiveProfile(wf, cfg)
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
