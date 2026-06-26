package main

import (
	"os/exec"
	"strings"

	"github.com/samuelnp/centinela/internal/workflow"
)

// gitOwner returns the latest committer name on the feature's branch, or
// "unknown" on any error / no commits. It is a package-level var so tests can
// override it without touching real git. Best-effort: a flaky or absent git
// never fails the dashboard — the owner column is advisory.
var gitOwner = func(repoRoot, feature string) string {
	cmd := exec.Command("git", "log", "-1", "--format=%an", feature)
	if repoRoot != "" {
		cmd.Dir = repoRoot
	}
	out, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	name := strings.TrimSpace(string(out))
	if name == "" {
		return "unknown"
	}
	return name
}

// dashboardOwners derives a feature->owner map over the active workflows using
// the gitOwner seam. The feature slug doubles as the branch name; any git error
// resolves to "unknown" inside gitOwner.
func dashboardOwners(active []*workflow.Workflow) map[string]string {
	owners := make(map[string]string, len(active))
	for _, wf := range active {
		if wf == nil {
			continue
		}
		owners[wf.Feature] = gitOwner("", wf.Feature)
	}
	return owners
}
