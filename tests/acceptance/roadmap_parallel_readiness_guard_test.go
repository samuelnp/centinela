package acceptance_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func seedWorkflowAt(t *testing.T, dir, feature, step string) {
	t.Helper()
	body := `{"feature":"` + feature + `","currentStep":"` + step + `","steps":{}}`
	if err := os.WriteFile(filepath.Join(dir, ".workflow", feature+".json"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// greenfieldDir builds a full greenfield fixture (PROJECT.md + roadmap with a
// Phase 0 bootstrap + analysis + quality artifacts) so `centinela start` reaches
// the dependency guard. names are all Phase 1 features; bootstrapDone marks setup.
func greenfieldDir(t *testing.T, roadmapJSON string, names []string, bootstrapDone bool) string {
	t.Helper()
	d := acceptanceDir(t, roadmapJSON)
	write := func(rel, body string) {
		if err := os.WriteFile(filepath.Join(d, rel), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("PROJECT.md", "Project Stage: greenfield\n")
	all := append([]string{"setup"}, names...)
	write(".workflow/roadmap-analysis.md", "# ok")
	write(".workflow/roadmap-analysis.json", analysisJSON(all))
	write(".workflow/roadmap-quality.md", "# ok")
	write(".workflow/roadmap-quality.json", qualityJSON(all))
	if bootstrapDone {
		seedDoneAt(t, d, "setup")
	}
	return d
}

func analysisJSON(names []string) string {
	feats := make([]string, len(names))
	for i, n := range names {
		feats[i] = `{"name":"` + n + `"}`
	}
	return `{"role":"senior-product-manager","features":[` + strings.Join(feats, ",") + `]}`
}

func qualityJSON(names []string) string {
	feats := make([]string, len(names))
	for i, n := range names {
		feats[i] = `{"name":"` + n + `","summary":"ok","scores":{"acceptanceCriteria":9,` +
			`"userValue":9,"definitionClarity":9,"dependencies":9,"effortEstimation":9,"overall":9}}`
	}
	return `{"role":"roadmap-quality-evaluator","threshold":9,"features":[` + strings.Join(feats, ",") + `]}`
}

const phase0 = `{"name":"Phase 0: Bootstrap","features":[{"name":"setup"}]}`

// Acceptance: start is refused when a dependency is not done (planned) — the error
// names the unmet dep and the feature; refused when only in-progress; proceeds when
// the dep is done; proceeds when the feature has no dependencies.
func TestAcceptance_StartGuard(t *testing.T) {
	bin := buildCent(t)
	rm := `{"phases":[` + phase0 + `,{"name":"P1","features":[
		{"name":"feature-a"},{"name":"feature-b","dependsOn":["feature-a"]}]}]}`

	// feature-a planned → start feature-b refused, names both.
	dir := greenfieldDir(t, rm, []string{"feature-a", "feature-b"}, true)
	out, code := runCent(t, bin, dir, "start", "feature-b")
	if code == 0 {
		t.Fatalf("start blocked feature should fail, got exit 0:\n%s", out)
	}
	if !strings.Contains(out, "feature-a") || !strings.Contains(out, "feature-b") {
		t.Fatalf("error must name unmet dep feature-a and feature-b:\n%s", out)
	}

	// feature-a in-progress → still refused, names feature-a.
	dir = greenfieldDir(t, rm, []string{"feature-a", "feature-b"}, true)
	seedWorkflowAt(t, dir, "feature-a", "code")
	out, code = runCent(t, bin, dir, "start", "feature-b")
	if code == 0 || !strings.Contains(out, "feature-a") {
		t.Fatalf("in-progress dep should refuse + name feature-a (exit=%d):\n%s", code, out)
	}

	// feature-a done → start feature-b proceeds (no dependency error).
	dir = greenfieldDir(t, rm, []string{"feature-a", "feature-b"}, true)
	seedDoneAt(t, dir, "feature-a")
	out, _ = runCent(t, bin, dir, "start", "feature-b")
	if strings.Contains(out, "blocked by unmet dependencies") {
		t.Fatalf("done dep should NOT emit a dependency error:\n%s", out)
	}

	// feature with no deps proceeds.
	dir = greenfieldDir(t, rm, []string{"feature-a", "feature-b"}, true)
	out, _ = runCent(t, bin, dir, "start", "feature-a")
	if strings.Contains(out, "blocked by unmet dependencies") {
		t.Fatalf("no-dep feature should NOT emit a dependency error:\n%s", out)
	}
}

// Acceptance: load-time rejection — cycle (2-node + 3-node), self-dependency, and
// an unknown dependency slug all fail `roadmap ready` (which loads + validates).
func TestAcceptance_LoadRejections(t *testing.T) {
	bin := buildCent(t)
	cases := []struct {
		name, json, want string
	}{
		{"unknown", `{"phases":[{"name":"P","features":[{"name":"b","dependsOn":["ghost"]}]}]}`, "ghost"},
		{"self", `{"phases":[{"name":"P","features":[{"name":"a","dependsOn":["a"]}]}]}`, "cycle"},
		{"two-node", `{"phases":[{"name":"P","features":[
			{"name":"a","dependsOn":["b"]},{"name":"b","dependsOn":["a"]}]}]}`, "cycle"},
		{"three-node", `{"phases":[{"name":"P","features":[
			{"name":"a","dependsOn":["c"]},{"name":"b","dependsOn":["a"]},
			{"name":"c","dependsOn":["b"]}]}]}`, "cycle"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := acceptanceDir(t, tc.json)
			out, code := runCent(t, bin, dir, "roadmap", "ready")
			if code == 0 {
				t.Fatalf("%s graph must fail to load, got exit 0:\n%s", tc.name, out)
			}
			if !strings.Contains(out, tc.want) {
				t.Fatalf("%s error should mention %q:\n%s", tc.name, tc.want, out)
			}
		})
	}
}
