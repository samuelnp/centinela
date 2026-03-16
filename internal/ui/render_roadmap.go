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
		StyleYellow.Render("⚠  ROADMAP.md not found — roadmap required"),
		"",
		"PROJECT.md is configured. Do not answer the user's message.",
		"Instead, help them define the project roadmap:",
		"",
		StyleMuted.Render("1. Ask about the main phases and features they envision"),
		StyleMuted.Render("2. Propose a phased roadmap — iterate until the user approves"),
		StyleMuted.Render("3. Write ROADMAP.md (readable markdown with phases + features)"),
		StyleMuted.Render("4. Write .workflow/roadmap.json in this exact format:"),
		StyleMuted.Render(`   {"phases":[{"name":"Phase 1","features":[{"name":"feature-slug"}]}]}`),
		StyleMuted.Render("   Feature names must be valid centinela slugs (lowercase, hyphens)"),
		StyleMuted.Render("5. Suggest: centinela start <first-feature>"),
		"",
		StyleRed.Render("Do not start any feature until the roadmap is approved."),
	)
	return boxStyle.Render(body)
}

// RenderRoadmapSummary returns a compact one-line roadmap progress indicator.
func RenderRoadmapSummary(r *roadmap.Roadmap) string {
	planned, inProgress, done := r.Summary()
	total := planned + inProgress + done
	line := fmt.Sprintf("Roadmap: %d/%d done", done, total)
	if inProgress > 0 {
		line += fmt.Sprintf(" · %d in-progress", inProgress)
	}
	return StyleMuted.Render(line)
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
	return boxStyle.Render(strings.Join(sections, "\n\n"))
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
