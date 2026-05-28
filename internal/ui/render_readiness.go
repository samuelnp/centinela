package ui

import (
	"strings"

	"github.com/samuelnp/centinela/internal/roadmap"
)

// roadmapIcon returns a status icon for a given status string.
// Kept for internal test compatibility; readinessMarker is the preferred path.
func roadmapIcon(status string) string {
	switch status {
	case "done":
		return IconDone
	case "in-progress":
		return IconActive
	case "ready":
		return IconReady
	case "blocked":
		return IconBlocked
	default:
		return IconPending
	}
}

// readinessMarker returns the icon and status annotation for a feature line.
func readinessMarker(fr roadmap.FeatureReadiness) (icon, annotation string) {
	switch fr.State {
	case "done":
		return IconDone, "(done)"
	case "in-progress":
		return IconActive, "(in-progress)"
	case "ready":
		return IconReady, "(ready)"
	case "blocked":
		names := strings.Join(fr.BlockedBy, ", ")
		return IconBlocked, "(blocked-by: " + names + ")"
	default:
		return IconPending, "(planned)"
	}
}

// RenderReadyList returns a styled list of ready feature names, one per line.
func RenderReadyList(ready []string) string {
	if len(ready) == 0 {
		return StyleMuted.Render("(none ready to start right now)")
	}
	lines := make([]string, len(ready))
	for i, name := range ready {
		lines[i] = "  " + IconReady + " " + name
	}
	return strings.Join(lines, "\n")
}
