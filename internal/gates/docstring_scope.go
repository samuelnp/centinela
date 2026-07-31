package gates

import (
	"strings"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/docstring"
	"github.com/samuelnp/centinela/internal/gitdiff"
)

// docstringRatchet resolves the changed-file set when the run handed the gate
// no filter. It uses the same resolver, the same merge base and the same
// diff_base every other file-scoped gate uses — one notion of "changed", not
// two. Overridable in tests. Returns the set and a non-empty degrade reason.
var docstringRatchet = func(base string) (*gitdiff.Set, string) {
	set, summary, _ := gitdiff.Default.ChangedFiles(base, true)
	return set, summary.Degrade
}

// docstringFiles resolves the files to inspect, or a non-empty Skip message.
//
// A nil filter means the run asked for a whole-repo scan. This gate declines
// it: a full scan of a codebase with a documentation backlog is a *report*
// (`centinela docs lint --full`), not a gate, and a gate that is permanently
// red teaches everyone to ignore it. So the gate resolves the changed-file set
// itself and enforces there. If git cannot answer, the gate reports Skip with
// the reason — never a confident pass, and never a red main for legacy debt.
func docstringFiles(cfg *config.Config, filter *gitdiff.Set) ([]string, string) {
	if filter == nil {
		var degrade string
		filter, degrade = docstringRatchet(cfg.Validate.DiffBase)
		if degrade != "" {
			return nil, "Changed-file scope unresolved (" + degrade + ") — nothing inspected."
		}
	}
	if filter.Len() == 0 {
		return nil, "No changed files — nothing inspected."
	}
	opts := DocstringOptions(cfg)
	var files []string
	for _, p := range filter.Paths() {
		if docstring.InScope(p, opts) {
			files = append(files, p)
		}
	}
	if len(files) == 0 {
		return nil, docstringNoGoFilesReason(cfg)
	}
	return files, ""
}

// docstringNoGoFilesReason names the roots the scan was confined to. Saying a
// bare "No Go files in scope" when the changed set held Go files that simply
// sat outside the configured roots reports the wrong cause — the operator sees
// a green-ish skip and no hint that the roots are the reason.
func docstringNoGoFilesReason(cfg *config.Config) string {
	roots := cfg.Gates.Docstring.Roots
	if len(roots) == 0 {
		roots = docstring.DefaultRoots
	}
	return "No Go files in scope under roots [" + strings.Join(roots, " ") +
		"] — nothing inspected."
}
