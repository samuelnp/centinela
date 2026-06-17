package doctor

import (
	"os"

	"github.com/samuelnp/centinela/internal/roadmap"
)

// describeRoadmap composes the one-line roadmap message for the combination of
// phase-name glyphs and ROADMAP.md drift that was found.
func describeRoadmap(glyphs []string, drifted bool) string {
	switch {
	case len(glyphs) > 0 && drifted:
		return "phase name(s) contain a live-status glyph and ROADMAP.md is out of sync"
	case len(glyphs) > 0:
		return "phase name(s) contain a live-status glyph that breaks phase-prefix detection"
	default:
		return roadmapMarkdownFile + " is out of sync with roadmap.json"
	}
}

// writeRoadmapMarkdown renders ROADMAP.md from the in-memory roadmap and writes
// it to disk with the same encoding the roadmap_drift gate compares against.
func writeRoadmapMarkdown(rm *roadmap.Roadmap) error {
	return os.WriteFile(roadmapMarkdownFile, roadmap.RenderMarkdown(rm), 0644)
}
