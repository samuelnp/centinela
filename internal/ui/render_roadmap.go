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
		StyleMuted.Render(`   {"phases":[{"name":"Phase 0: Bootstrap","features":[{"name":"project-bootstrap"}]},{"name":"Phase 1","features":[{"name":"feature-slug"}]}]}`),
		StyleMuted.Render("   Feature names must be valid centinela slugs (lowercase, hyphens)"),
		StyleMuted.Render("   If PROJECT.md says Project Stage: existing, Phase 0 is optional"),
		StyleMuted.Render("5. For each Phase 1 feature, write docs/features/<slug>.md with:"),
		StyleMuted.Render("   ## Problem · ## User Stories · ## Acceptance Criteria"),
		StyleMuted.Render("   ## Edge Cases · ## Data Model · ## Integration Points"),
		StyleMuted.Render("   ## Risks · ## Decomposition"),
		StyleMuted.Render("6. Suggest: centinela start <first-feature>"),
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

// RenderRoadmap returns a full styled roadmap with per-feature status.
func RenderRoadmap(r *roadmap.Roadmap) string {
	var sections []string
	for _, phase := range r.Phases {
		lines := []string{StyleBold.Render(phase.Name)}
		for _, f := range phase.Features {
			status := roadmap.FeatureStatus(f.Name)
			icon := roadmapIcon(status)
			lines = append(lines, "  "+icon+" "+f.Name+
				StyleMuted.Render("  ("+status+")"))
		}
		sections = append(sections, strings.Join(lines, "\n"))
	}
	body := strings.Join(sections, "\n\n")
	return renderSystemPanel("ROADMAP", "PHASE OVERVIEW", toneInfo, body)
}

func roadmapIcon(status string) string {
	switch status {
	case "done":
		return IconDone
	case "in-progress":
		return IconActive
	default:
		return IconPending
	}
}
