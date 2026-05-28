package evidence

import (
	"fmt"
	"time"
)

// RenderTemplate returns the (path, body) pairs the writer should drop for a
// given kind. Pure — does no I/O. Most kinds emit one file; the
// documentation-specialist kind emits a JSON+MD pair.
func RenderTemplate(kind ArtifactKind, feature string) ([]string, [][]byte, error) {
	switch kind {
	case KindEdgeCases:
		return single(artifactPath(feature, "edge-cases.md"), edgeCasesBody(feature))
	case KindGatekeeper:
		return single(artifactPath(feature, "gatekeeper.md"), gatekeeperBody(feature))
	case KindProductionReadiness:
		return single(artifactPath(feature, "production-readiness.md"), prodReadyBody(feature))
	case KindDocumentationSpecialist:
		return docsSpecialistPair(feature)
	default:
		return nil, nil, fmt.Errorf("unsupported artifact kind %q", kind)
	}
}

// today returns the YYYY-MM-DD stamp used in markdown templates. Extracted
// so tests can stay deterministic via a single source of "now".
func today() string {
	return time.Now().UTC().Format("2006-01-02")
}

func single(path string, body []byte) ([]string, [][]byte, error) {
	return []string{path}, [][]byte{body}, nil
}
