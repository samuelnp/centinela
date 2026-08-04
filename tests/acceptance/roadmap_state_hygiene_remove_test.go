// Acceptance: specs/roadmap-state-hygiene.feature
package acceptance_test

import (
	"strings"
	"testing"
)

// Scenario: the removal message no longer tells the operator to sync by hand
func TestRsh_RemoveMessageDoesNotMentionManualSync(t *testing.T) {
	bin := buildCent(t)
	dir := rshRepo(t, rshBaseRoadmap)

	out, code := runCent(t, bin, dir, "roadmap", "remove", "feature-c")
	if code != 0 {
		t.Fatalf("remove exit=%d\n%s", code, out)
	}
	if strings.Contains(out, "Remember to sync") {
		t.Fatalf("output must not tell the operator to sync by hand:\n%s", out)
	}
}
