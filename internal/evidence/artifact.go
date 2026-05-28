package evidence

import (
	"fmt"
	"path/filepath"
	"sort"

	"github.com/samuelnp/centinela/internal/workflow"
)

// ArtifactKind identifies a templated evidence artifact distinct from the
// per-role JSON pair (which is owned by `centinela evidence init`).
// Artifact templates are pre-filled stubs an agent can extend.
type ArtifactKind string

const (
	// KindEdgeCases is the .workflow/<feature>-edge-cases.md companion
	// required by the tests step.
	KindEdgeCases ArtifactKind = "edge-cases"
	// KindGatekeeper is the validate-step gatekeeper report.
	KindGatekeeper ArtifactKind = "gatekeeper"
	// KindProductionReadiness is the validate-step PR review report.
	KindProductionReadiness ArtifactKind = "production-readiness"
	// KindDocumentationSpecialist is the docs-step JSON+MD pair.
	KindDocumentationSpecialist ArtifactKind = "documentation-specialist"
)

// KindsAllowed lists every supported artifact kind in stable order for
// error messages and CLI help.
func KindsAllowed() []ArtifactKind {
	return []ArtifactKind{
		KindEdgeCases,
		KindGatekeeper,
		KindProductionReadiness,
		KindDocumentationSpecialist,
	}
}

// ParseKind validates an arbitrary string against KindsAllowed.
func ParseKind(s string) (ArtifactKind, error) {
	for _, k := range KindsAllowed() {
		if string(k) == s {
			return k, nil
		}
	}
	allowed := make([]string, 0, len(KindsAllowed()))
	for _, k := range KindsAllowed() {
		allowed = append(allowed, string(k))
	}
	sort.Strings(allowed)
	return "", fmt.Errorf("unknown artifact kind %q (allowed: %v)", s, allowed)
}

// artifactPath returns the .workflow/<feature>-<suffix> path. Suffix
// includes the file extension.
func artifactPath(feature, suffix string) string {
	return filepath.Join(workflow.WorkflowDir, feature+"-"+suffix)
}
