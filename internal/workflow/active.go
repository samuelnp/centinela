package workflow

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ActiveWorkflows scans dir for genuine workflow-state files and returns the
// ones that are active (a real, non-empty, non-"done" step). A file qualifies
// ONLY when its parsed Feature equals the file's base name (<feature>.json) and
// CurrentStep is a real, non-empty, non-"done" step. This rejects per-role
// evidence JSONs (<feature>-<role>.json, empty CurrentStep) and ad-hoc roadmap
// JSONs (roadmap.json, roadmap-quality.json). Survivors are deduped by Feature
// (most-recently-touched wins) and sorted by file mtime descending.
func ActiveWorkflows(dir string) []*Workflow {
	entries, _ := filepath.Glob(filepath.Join(dir, "*.json")) //nolint:errcheck
	type tracked struct {
		wf    *Workflow
		mtime int64
	}
	byFeature := map[string]tracked{}
	for _, p := range entries {
		base := strings.TrimSuffix(filepath.Base(p), ".json")
		wf, err := Load(base)
		if err != nil || wf.Feature != base {
			continue
		}
		if wf.CurrentStep == "" || wf.CurrentStep == "done" {
			continue
		}
		info, err := os.Stat(p)
		if err != nil {
			continue
		}
		m := info.ModTime().UnixNano()
		if prev, ok := byFeature[wf.Feature]; ok && prev.mtime >= m {
			continue
		}
		byFeature[wf.Feature] = tracked{wf: wf, mtime: m}
	}
	out := make([]tracked, 0, len(byFeature))
	for _, t := range byFeature {
		out = append(out, t)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].mtime > out[j].mtime })
	result := make([]*Workflow, 0, len(out))
	for _, t := range out {
		result = append(result, t.wf)
	}
	return result
}

// CapActive returns at most max workflows (the front of wfs) and the count of
// workflows omitted beyond the cap. max <= 0 means no cap.
func CapActive(wfs []*Workflow, max int) (shown []*Workflow, more int) {
	if max <= 0 || len(wfs) <= max {
		return wfs, 0
	}
	return wfs[:max], len(wfs) - max
}
