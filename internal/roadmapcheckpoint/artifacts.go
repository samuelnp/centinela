package roadmapcheckpoint

import "time"

// RequiredArtifacts returns the canonical list of roadmap-defining files
// whose mtimes determine marker freshness. Order is not significant.
func RequiredArtifacts() []string {
	return []string{
		"ROADMAP.md",
		".workflow/roadmap.json",
		".workflow/roadmap-analysis.md",
		".workflow/roadmap-analysis.json",
		".workflow/roadmap-quality.md",
		".workflow/roadmap-quality.json",
	}
}

// LatestMtime returns the most recent modification time across paths. The
// second return value is false when none of the paths exist (callers can
// then choose to suppress rather than emit a meaningless decision).
func LatestMtime(paths []string, fs FS) (time.Time, bool) {
	var latest time.Time
	found := false
	for _, p := range paths {
		mt, ok := fs.Stat(p)
		if !ok {
			continue
		}
		if !found || mt.After(latest) {
			latest = mt
			found = true
		}
	}
	return latest, found
}
