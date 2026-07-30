// Acceptance: specs/token-diet.feature
package acceptance_test

import (
	"testing"
)

// Scenario: Path normalization is symmetric for briefs and plans
func TestTD_PathNormalizationSymmetricForBriefsAndPlans(t *testing.T) {
	bin := tdBuildBin(t)
	cases := []struct {
		written, other string
	}{
		{"docs/features/token-diet.md", "docs/plans/token-diet.md"},
		{"./docs/features/token-diet.md", "docs/plans/token-diet.md"},
		{`docs\features\token-diet.md`, "docs/plans/token-diet.md"},
		{"/home/user/repo/docs/features/token-diet.md", "docs/plans/token-diet.md"},
		{"docs/plans/token-diet.md", "docs/features/token-diet.md"},
		{"./docs/plans/token-diet.md", "docs/features/token-diet.md"},
		{`docs\plans\token-diet.md`, "docs/features/token-diet.md"},
		{"/home/user/repo/docs/plans/token-diet.md", "docs/features/token-diet.md"},
	}
	for _, tc := range cases {
		dir := tdRepo(t)
		tdFeatureDocs(t, dir, "token-diet")
		tdWritePlannerEvidence(t, dir, "token-diet", []string{tc.written, tc.other})
		out, code := runCent(t, bin, dir, "evidence", "validate", "token-diet")
		if code != 0 {
			t.Fatalf("written=%q: expected validate to pass, got exit %d: %s", tc.written, code, out)
		}
	}
}
