// Acceptance: specs/token-diet.feature
package acceptance_test

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const tdRoadmapLine = "Roadmap: "

// tdRoadmap1 and tdRoadmap2 are two DIFFERENT stable projections (different
// feature membership => different SummaryDigest) for the reformat/mutation
// scenarios.
const tdRoadmap1 = `{"phases":[{"name":"P1","features":[{"name":"widget-a"}]}]}`
const tdRoadmap2 = `{"phases":[{"name":"P1","features":[{"name":"widget-a"},{"name":"widget-b"}]}]}`

// tdRoadmap1Reformatted is tdRoadmap1 with different whitespace/key order but
// an identical projection (same phase name, same feature name+status).
const tdRoadmap1Reformatted = `{
  "phases": [
    {
      "features": [ { "name": "widget-a" } ],
      "name": "P1"
    }
  ]
}`

// tdDigestPath is the digest state file's path relative to a repo dir.
func tdDigestPath(dir string) string {
	return filepath.Join(dir, ".workflow", ".roadmap-digest")
}

// tdWriteRoadmap writes .workflow/roadmap.json at dir.
func tdWriteRoadmap(t *testing.T, dir, json string) {
	t.Helper()
	mustWrite(t, filepath.Join(dir, ".workflow", "roadmap.json"), json)
}

// tdHookContext runs `hook context` with the given stdin body and returns
// combined output + exit code. A nil/absent body-arg distinction matters for
// the fail-open scenarios, so the caller passes the exact bytes to send.
func tdHookContext(t *testing.T, bin, dir, stdin string) (string, int) {
	t.Helper()
	cmd := exec.Command(bin, "hook", "context")
	cmd.Dir = dir
	cmd.Stdin = strings.NewReader(stdin)
	out, err := cmd.CombinedOutput()
	code := 0
	if ee, ok := err.(*exec.ExitError); ok {
		code = ee.ExitCode()
	} else if err != nil {
		t.Fatalf("run hook context: %v", err)
	}
	return string(out), code
}

// tdSessionPayload renders a minimal UserPromptSubmit-shaped stdin body.
func tdSessionPayload(sessionID string) string {
	return `{"session_id":"` + sessionID + `"}`
}
