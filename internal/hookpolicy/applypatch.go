package hookpolicy

import (
	"path/filepath"
	"strings"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/workflow"
)

// applyPatchPrefixes maps an apply_patch envelope line prefix to the path that
// follows it. Codex sends the patch as a single string in tool_input.command;
// file_path/filePath are absent, so the target paths must be parsed out here.
var applyPatchPrefixes = []string{
	"*** Add File: ",
	"*** Update File: ",
	"*** Delete File: ",
	"*** Move to: ",
}

// ExtractApplyPatchPaths scans an apply_patch envelope and returns every file
// path it touches (Add/Update/Delete/Move). Returns nil when none are found.
func ExtractApplyPatchPaths(command string) []string {
	var paths []string
	for _, line := range strings.Split(command, "\n") {
		line = strings.TrimSpace(line)
		for _, p := range applyPatchPrefixes {
			if strings.HasPrefix(line, p) {
				if path := strings.TrimSpace(strings.TrimPrefix(line, p)); path != "" {
					paths = append(paths, path)
				}
				break
			}
		}
	}
	return paths
}

// EvaluatePrewriteMulti evaluates each path via EvaluatePrewrite and returns the
// first blocking (non-Allow) decision. If every path is allowed it returns an
// Allow decision. With no paths it returns Allow (unchanged no-op behavior).
func EvaluatePrewriteMulti(paths []string, cwd string, cfg *config.Config, wfs []*workflow.Workflow) PrewriteDecision {
	for _, path := range paths {
		// Codex apply_patch paths are repo-relative; resolve against cwd so
		// isInsideWorkspace/ClassifyFile see an absolute path. Claude/OpenCode
		// already send absolute file_path, so IsAbs short-circuits unchanged.
		abs := path
		if cwd != "" && !filepath.IsAbs(abs) {
			abs = filepath.Join(cwd, abs)
		}
		d := EvaluatePrewrite(abs, cwd, cfg, wfs)
		if !d.Allow {
			d.Path = path // report the original path Codex gave, for rendering
			return d
		}
	}
	return PrewriteDecision{Allow: true}
}
