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
	out := make([]FixHint, 0, len(errs)+1)
	for _, fe := range errs {
		out = append(out, fixHintFor(feature, role, fe))
	}
	if hint, ok := handoffChainHint(feature, role, r); ok {
		out = append(out, hint)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// handoffChainHint asks the completion gate's OWN chain question, so the
// pre-flight command the role prompts tell agents to run cannot report
// "evidence ok" for evidence `centinela complete` will refuse. An agent's
// self-check contradicting the gate is a false success one layer up.
//
// Scope is deliberately identical to the gate's — only the roles that step
// actually requires — so this can never be STRICTER than what it previews.
// That excludes the out-of-band merge-steward (whose literal verdict pair is
// checked in ValidateEvidence) and the optional production-readiness role.
func handoffChainHint(feature string, role Role, r *RoleEvidence) (FixHint, bool) {
	required := false
	for _, want := range workflow.RequiredEvidenceRoles(feature, r.Step) {
		required = required || want == role
	}
	if !required {
		return FixHint{}, false
	}
	err := workflow.CheckHandoffTo(feature, r.Step, role, r.HandoffTo)
	if err == nil {
		return FixHint{}, false
	}
	return FixHint{Feature: feature, Role: role, Field: "handoffTo", Message: err.Error()}, true
}
