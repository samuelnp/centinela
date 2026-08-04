package workflow

import (
	"strings"

	"github.com/samuelnp/centinela/internal/roadmapstate"
	"github.com/samuelnp/centinela/internal/treestate"
)

// revisionRangeIsRoadmapStateOnly reports whether every path that changed
// between the stamped revision and HEAD is roadmap state.
//
// This is the committed half of the D5 exemption: a mutation now moves HEAD by
// itself, so a verifier that records a deferral after stamping would otherwise
// void its own verdict. It is a strict subset test (roadmapstate.Covers), NOT a
// filter — a range mixing roadmap state with any other path is not exempt, so a
// real change hidden in the same range still stales the stamp (R1).
//
// It fails CLOSED: a git error, an unresolvable recorded revision, or anything
// it cannot read reports false, and the caller reports the verification stale.
func revisionRangeIsRoadmapStateOnly(root, recorded string, run treestate.Runner) bool {
	if strings.TrimSpace(recorded) == "" {
		return false
	}
	out, err := run(root, "diff", "--name-only", recorded+"..HEAD")
	if err != nil {
		return false
	}
	paths := nonEmptyLines(out)
	// An empty range means the two revisions have identical trees, so the
	// stamped tree state still holds — the strongest possible evidence, not a
	// missing answer.
	if len(paths) == 0 {
		return true
	}
	return roadmapstate.Covers(paths)
}

// nonEmptyLines splits git's newline-separated path list, dropping blanks.
func nonEmptyLines(out string) []string {
	var paths []string
	for _, ln := range strings.Split(out, "\n") {
		if s := strings.TrimSpace(ln); s != "" {
			paths = append(paths, s)
		}
	}
	return paths
}
