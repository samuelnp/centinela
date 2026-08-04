// Package roadmapstate owns THE definition of what "roadmap state" is: the
// governance bookkeeping files a roadmap mutation rewrites, as opposed to the
// product the repository ships.
//
// One list drives three consumers that must never disagree — the pathspec a
// mutation's auto-commit uses (internal/roadmap), the working-tree digest
// exclusion (internal/treestate) and the freshness revision-range exemption
// (internal/workflow). Owning it here rather than in any of them is what keeps
// treestate -> roadmap -> workflow -> treestate from becoming an import cycle,
// the same reasoning as internal/acceptance and internal/worktreepath.
//
// It imports the standard library only.
package roadmapstate

import "strings"

const (
	// RoadmapJSON is the canonical machine-readable roadmap.
	RoadmapJSON = ".workflow/roadmap.json"
	// MarkdownFile is the human-readable roadmap generated from RoadmapJSON.
	MarkdownFile = "ROADMAP.md"
	// WorkflowDir is verification's own output directory. Everything under it
	// is governance bookkeeping (D3a: the verifier's report must not stale the
	// stamp it is about to write), so it counts as roadmap state too.
	WorkflowDir = ".workflow/"
)

// Paths returns the roadmap-state files EVERY mutation rewrites, in commit
// order. A mutation that also rewrites artifacts (promote) supplies those as
// extra paths; the always-present pair lives here so no caller hardcodes it.
func Paths() []string {
	return []string{RoadmapJSON, MarkdownFile}
}

// IsStatePath reports whether p denotes roadmap state. Paths are normalized
// first: git quotes paths containing unusual bytes and pads porcelain fields,
// so trimming here keeps every caller from repeating it.
func IsStatePath(p string) bool {
	p = normalize(p)
	if p == "" {
		return false
	}
	return p == MarkdownFile || strings.HasPrefix(p, WorkflowDir)
}

// Covers reports whether EVERY path is roadmap state — a strict subset test,
// never a filter. A mixed set (roadmap state plus a source file) is NOT
// covered, which is what stops the freshness exemption from becoming a hole:
// a real change hidden alongside roadmap churn still stales the stamp.
//
// An empty set is deliberately NOT covered. "Nothing changed" is a different
// question from "only roadmap state changed", and the callers that want the
// empty case to pass say so explicitly.
func Covers(paths []string) bool {
	if len(paths) == 0 {
		return false
	}
	for _, p := range paths {
		if !IsStatePath(p) {
			return false
		}
	}
	return true
}

func normalize(p string) string {
	return strings.Trim(strings.TrimSpace(p), `"`)
}
