package evidence

import (
	"fmt"
	"os"
)

// Invalidate removes the .workflow/<feature>-<role>.{json,md} certification
// pair so a re-opened step's evidence reads as missing and the next
// `centinela complete` is forced to re-run that role's subagent. A missing
// file is NOT an error (idempotent); it reports whether anything was removed.
// It touches ONLY the role evidence pair under .workflow/ — never docs/,
// tests/, or source, which are work product, not certification.
func Invalidate(feature string, role Role) (bool, error) {
	return removeBoth(pathFor(feature, role), companionPath(feature, role))
}

// InvalidateArtifact removes a non-role .workflow/<feature>-<suffix> artifact
// (e.g. "edge-cases.md"). Suffix includes the extension. Missing is not an
// error; it reports whether the file was removed.
func InvalidateArtifact(feature, suffix string) (bool, error) {
	return removeBoth(artifactPath(feature, suffix))
}

// removeBoth deletes each path, treating absence as success. It returns true
// when at least one path existed and was removed, and surfaces any non-absence
// removal error naming the offending path.
func removeBoth(paths ...string) (bool, error) {
	removed := false
	for _, p := range paths {
		err := os.Remove(p)
		switch {
		case err == nil:
			removed = true
		case os.IsNotExist(err):
			// idempotent: already gone.
		default:
			return removed, fmt.Errorf("invalidate %s: %w", p, err)
		}
	}
	return removed, nil
}
