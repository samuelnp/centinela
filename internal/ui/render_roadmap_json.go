package ui

import (
	"os"

	"github.com/charmbracelet/lipgloss"
)

func RenderRoadmapJSONNeeded(err error) string {
	status := "⚠ Machine-readable roadmap missing — .workflow/roadmap.json required"
	if !os.IsNotExist(err) {
		status = "⚠ Machine-readable roadmap invalid — fix .workflow/roadmap.json"
	}
	body := lipgloss.JoinVertical(lipgloss.Left, status, "", "ROADMAP.md exists. Do not answer the user's message.", "Instead, sync the machine-readable roadmap file:", "", StyleMuted.Render("1. Read ROADMAP.md and copy its phases and feature slugs exactly"), StyleMuted.Render("2. Write .workflow/roadmap.json in this exact format:"), StyleMuted.Render(`   {"phases":[{"name":"Phase 0: Bootstrap","features":[{"name":"project-bootstrap"}]},{"name":"Phase 1","features":[{"name":"feature-slug"}]}]}`), StyleMuted.Render("3. Keep every feature name aligned with ROADMAP.md"), StyleMuted.Render("4. Run centinela roadmap validate once JSON, analysis, and quality artifacts exist"), StyleMuted.Render("5. See docs/architecture/artifact-templates.md for the full setup and workflow file templates"), "", StyleRed.Render("Do not start any feature until .workflow/roadmap.json is valid."))
	return renderSystemPanel("SETUP", "ROADMAP JSON REQUIRED", toneWarn, body)
}
