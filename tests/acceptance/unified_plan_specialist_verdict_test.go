// Acceptance: specs/unified-plan-specialist.feature
package acceptance_test

import (
	"encoding/json"
	"sort"
	"strings"
	"testing"
)

// upsVerdictRoles runs `centinela verdict <feature>` and returns the sorted set
// of roles its claim-verification section actually checked. The command exits
// non-zero when any claim fails, which is expected here — the assertion is about
// WHICH roles were verified, not whether they passed.
func upsVerdictRoles(t *testing.T, bin, dir, feature string) []string {
	t.Helper()
	out, _ := runCent(t, bin, dir, "verdict", feature)
	start := strings.Index(out, "{")
	if start < 0 {
		t.Fatalf("verdict emitted no JSON packet: %s", out)
	}
	var pkt struct {
		Verify []struct {
			Role  string `json:"role"`
			Claim string `json:"claim"`
		} `json:"verify"`
	}
	// Decode (not Unmarshal): the packet is followed by the CLI's error line,
	// so the stream has trailing non-JSON content after the object.
	if err := json.NewDecoder(strings.NewReader(out[start:])).Decode(&pkt); err != nil {
		t.Fatalf("decode verdict packet: %v\n%s", err, out)
	}
	seen := map[string]bool{}
	var roles []string
	for _, c := range pkt.Verify {
		if c.Claim == "claims" {
			t.Fatalf("%s: verdict reported the no-claims placeholder despite real evidence: %s",
				feature, out)
		}
		if c.Role != "" && !seen[c.Role] {
			seen[c.Role] = true
			roles = append(roles, c.Role)
		}
	}
	sort.Strings(roles)
	return roles
}

// Scenario: Every claim-verification surface resolves the same required role set
//
// `centinela verdict` builds its own verify.Deps. Omitting Deps.Roles there made
// it fall back to the contract-blind policy layer, so a legacy plan workflow
// reported "no claims to verify" while `centinela verify` checked the same tree
// in full — a direct refutation of D3 on a shipped surface.
func TestUPS_VerdictSurfaceUsesContractAwareRoles(t *testing.T) {
	bin := upsBuildBin(t)
	cases := []struct {
		feature, contract string
		want              []string
	}{
		{"pinned-workflow", "planner-v1", []string{"planner"}},
		{"unpinned-workflow", "", []string{"big-thinker", "feature-specialist"}},
	}
	for _, tc := range cases {
		dir := upsExistingRepo(t)
		upsWriteWorkflow(t, dir, tc.feature, tc.contract)
		upsFeatureDocs(t, dir, tc.feature)
		for _, role := range tc.want {
			upsWriteEvidence(t, dir, tc.feature, role, "senior-engineer", []string{"an edge case"})
		}

		got := upsVerdictRoles(t, bin, dir, tc.feature)
		want := append([]string(nil), tc.want...)
		sort.Strings(want)
		if strings.Join(got, ",") != strings.Join(want, ",") {
			t.Fatalf("%s: verdict verified roles %v, want exactly %v", tc.feature, got, want)
		}
	}
}
