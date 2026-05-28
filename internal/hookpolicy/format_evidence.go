package hookpolicy

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"strings"

	"github.com/samuelnp/centinela/internal/workflow"
)

// FormatEvidence pretty-prints active-feature `.workflow/*.json` files in
// place. It is a pure function — callers feed in the path, body, and the
// active feature slug; the file is rewritten by the caller.
//
// Returns (out, changed, err):
//   - changed=false: path is out of scope, body is non-JSON, or the file
//     already matches the canonical output (idempotent passthrough). Out
//     equals the input body byte-for-byte.
//   - changed=true: body was reformatted; caller should rewrite the file
//     atomically with the returned bytes.
//
// Schema validity is NOT enforced here — `centinela evidence validate` is
// the gate. Parse failures fall through silently so unrelated JSON files
// the agent dropped in `.workflow/` are not destroyed.
func FormatEvidence(path string, body []byte, activeFeature string) ([]byte, bool, error) {
	if !isActiveFeatureEvidence(path, activeFeature) {
		return body, false, nil
	}
	pretty, ok := reformatJSON(body)
	if !ok {
		return body, false, nil
	}
	if bytes.Equal(pretty, body) {
		return body, false, nil
	}
	return pretty, true, nil
}

// isActiveFeatureEvidence reports whether path is `.workflow/<feature>-*.json`
// for the active feature. Other features' files in `.workflow/` and any
// non-JSON paths are out of scope.
func isActiveFeatureEvidence(path, activeFeature string) bool {
	if activeFeature == "" {
		return false
	}
	if filepath.Ext(path) != ".json" {
		return false
	}
	dir := filepath.ToSlash(filepath.Dir(path))
	if !strings.HasSuffix(dir, workflow.WorkflowDir) {
		return false
	}
	base := filepath.Base(path)
	return strings.HasPrefix(base, activeFeature+"-")
}

// reformatJSON re-encodes body with two-space indent and the canonical
// evidence key order. Returns ok=false if body is not a JSON object.
// json.Marshal/json.Indent on round-tripped bytes cannot fail at runtime
// (we are re-emitting what we just parsed), so only the parse step is
// guarded.
func reformatJSON(body []byte) ([]byte, bool) {
	var generic map[string]json.RawMessage
	if err := json.Unmarshal(body, &generic); err != nil {
		return nil, false
	}
	ordered := encodeOrderedEvidence(generic)
	var pretty bytes.Buffer
	_ = json.Indent(&pretty, ordered, "", "  ")
	pretty.WriteByte('\n')
	return pretty.Bytes(), true
}
