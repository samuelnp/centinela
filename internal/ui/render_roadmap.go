package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/samuelnp/centinela/internal/roadmap"
)

// RenderRoadmapNeeded returns context to inject when ROADMAP.md is missing.
func RenderRoadmapNeeded() string {
	body := lipgloss.JoinVertical(lipgloss.Left,
		StyleYellow.Render("⚠ ROADMAP.md not found — roadmap required"),
		"",
		"PROJECT.md is configured. Do not answer the user's message.",
		"Instead, help them define the project roadmap:",
		"",
		StyleMuted.Render("1. For each feature they mention, ask:"),
		StyleMuted.Render("   · What problem does it solve? Who is the user?"),
		StyleMuted.Render("   · Is it large enough to split into sub-features?"),
		StyleMuted.Render("2. Propose a phased roadmap — iterate until the user approves"),
		StyleMuted.Render("3. Write ROADMAP.md (readable markdown with phases + features)"),
		StyleMuted.Render("4. Write .workflow/roadmap.json in this exact format:"),
		StyleMuted.Render(`   {"phases":[{"name":"Phase 0: Bootstrap","features":[{"name":"project-bootstrap","dependsOn":[]}]},{"name":"Phase 1","features":[{"name":"feature-slug","dependsOn":["project-bootstrap"]}]}]}`),
		StyleMuted.Render("   Feature names must be valid centinela slugs (lowercase, hyphens)"),
		StyleMuted.Render("   dependsOn is optional; omit or set [] for no dependencies"),
		StyleMuted.Render("   If PROJECT.md says Project Stage: existing, Phase 0 is optional"),
		StyleMuted.Render("5. For each Phase 1 feature, write docs/features/<slug>.md with:"),
		StyleMuted.Render("   ## Problem · ## User Stories · ## Acceptance Criteria"),
		StyleMuted.Render("   ## Edge Cases · ## Data Model · ## Integration Points"),
		StyleMuted.Render("   ## Risks · ## Decomposition"),
		StyleMuted.Render("6. Delegate senior PM analysis and write roadmap artifacts:"),
		StyleMuted.Render("   .workflow/roadmap-analysis.md + .workflow/roadmap-analysis.json"),
		StyleMuted.Render("   JSON role must be senior-product-manager"),
		StyleMuted.Render("7. Delegate roadmap quality scoring and write quality artifacts:"),
		StyleMuted.Render("   .workflow/roadmap-quality.md + .workflow/roadmap-quality.json"),
		StyleMuted.Render("   Role roadmap-quality-evaluator, threshold 9, all features overall >= 9"),
		StyleMuted.Render("8. Suggest: centinela start <first-feature>"),
		"",
		StyleRed.Render("Do not start any feature until the roadmap AND feature briefs are written."),
	)
	return renderSystemPanel("SETUP", "ROADMAP REQUIRED", toneWarn, body)
}

// RenderRoadmapSummary returns a compact one-line roadmap progress indicator.
func RenderRoadmapSummary(r *roadmap.Roadmap) string {
	planned, inProgress, done := r.Summary()
	total := planned + inProgress + done
	line := fmt.Sprintf("Roadmap: %d/%d done", done, total)
	if inProgress > 0 {
		line += fmt.Sprintf(" · %d in-progress", inProgress)
	}
	return renderSystemLine("ROADMAP", line, toneInfo)
}

// RenderRoadmap returns a full styled roadmap with per-feature readiness status.
func RenderRoadmap(r *roadmap.Roadmap) string {
	readiness := roadmap.DeriveReadiness(r)
	idx := map[string]roadmap.FeatureReadiness{}
	for _, fr := range readiness {
		idx[fr.Name] = fr
	}
	var sections []string
	for _, phase := range r.Phases {
		if roadmap.IsBacklogPhaseName(phase.Name) {
			continue // Backlog findings render in their own section below
		}
		lines := []string{StyleBold.Render(phase.Name)}
		for _, f := range phase.Features {
			fr := idx[f.Name]
			icon, annotation := readinessMarker(fr)
			lines = append(lines, "  "+icon+" "+f.Name+
				StyleMuted.Render("  "+annotation))
		}
		sections = append(sections, strings.Join(lines, "\n"))
	}
	if backlog := renderBacklogSection(r); backlog != "" {
		sections = append(sections, backlog)
	}
	body := strings.Join(sections, "\n\n")
	return renderSystemPanel("ROADMAP", "PHASE OVERVIEW", toneInfo, body)
}
