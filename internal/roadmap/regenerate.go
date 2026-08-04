package roadmap

import (
	"bytes"
	"os"

	"github.com/samuelnp/centinela/internal/roadmapstate"
)

// RegenerateMarkdown rewrites ROADMAP.md from roadmap.json through the SAME
// renderer the roadmap_drift gate compares against, and reports whether the
// bytes actually changed. Writing only on a real change is what keeps a no-op
// mutation from producing an empty commit.
func RegenerateMarkdown() (bool, error) {
	r, err := Load()
	if err != nil {
		return false, err
	}
	data := RenderMarkdown(r)
	if old, err := os.ReadFile(roadmapstate.MarkdownFile); err == nil && bytes.Equal(old, data) {
		return false, nil
	}
	return true, writeAtomic(roadmapstate.MarkdownFile, data)
}

// PromoteArtifactPaths returns the artifact files a scored promote rewrites in
// addition to the always-present roadmap-state pair. A mutation that writes a
// file outside its declared pathspec is the exact footgun this feature exists
// to kill, so this list is asserted in tests against what promote touches.
func PromoteArtifactPaths() []string {
	return []string{
		RoadmapAnalysisFile, RoadmapQualityFile,
		RoadmapAnalysisMarkdown, RoadmapQualityMarkdown,
	}
}

// WriteRoadmapJSON atomically replaces roadmap.json with data. It is the write
// half of `roadmap resolve`: the merged document is produced entirely in
// memory first, so a refused merge leaves the operator's conflict markers on
// disk byte-identical.
func WriteRoadmapJSON(data []byte) error {
	return writeAtomic(RoadmapFile, data)
}
