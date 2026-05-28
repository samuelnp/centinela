package evidence

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/samuelnp/centinela/internal/workflow"
)

// ValidateFeature walks every .workflow/<feature>-*.json artifact and emits
// one FixHint per validation failure. UIPaths are consulted for the
// ux-ui-specialist outputs rule; pass nil if the feature is not user-facing.
func ValidateFeature(feature string, uiPaths []string) []FixHint {
	prefix := feature + "-"
	matches, _ := filepath.Glob(filepath.Join(workflow.WorkflowDir, prefix+"*.json"))
	hints := []FixHint{}
	for _, path := range matches {
		role, ok := roleFromPath(path, feature)
		if !ok {
			continue
		}
		hints = append(hints, hintsForFile(feature, role, path, uiPaths)...)
	}
	return hints
}

func roleFromPath(path, feature string) (Role, bool) {
	base := strings.TrimSuffix(filepath.Base(path), ".json")
	rest := strings.TrimPrefix(base, feature+"-")
	if rest == "" || rest == base {
		return "", false
	}
	if !IsKnownRole(Role(rest)) {
		return "", false
	}
	return Role(rest), true
}

func hintsForFile(feature string, role Role, path string, uiPaths []string) []FixHint {
	r, err := Read(feature, role)
	if err != nil {
		return []FixHint{{
			Feature: feature, Role: role,
			Message: fmt.Sprintf("read failed: %v", err),
			Command: fmt.Sprintf("centinela evidence init %s %s", feature, role),
		}}
	}
	errs := r.Validate(path, uiPaths)
	if len(errs) == 0 {
		return nil
	}
	out := make([]FixHint, 0, len(errs))
	for _, fe := range errs {
		out = append(out, fixHintFor(feature, role, fe))
	}
	return out
}
