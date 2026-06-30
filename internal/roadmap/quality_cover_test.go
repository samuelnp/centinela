package roadmap

import (
	"os"
	"strings"
	"testing"
)

const goodScores = `"scores":{"acceptanceCriteria":9,"userValue":9,"definitionClarity":9,"dependencies":9,"effortEstimation":9,"overall":9}`

// quality wraps a features array body in the canonical report envelope.
func qualityReport(features string) string {
	return `{"role":"roadmap-quality-evaluator","threshold":9,"features":[` + features + `]}`
}

// ValidateQuality covers the per-feature loop error branches.
func TestValidateQuality_FeatureLoopBranches(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := os.MkdirAll(".workflow", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(RoadmapQualityMarkdown, []byte("# q"), 0644); err != nil {
		t.Fatal(err)
	}
	r := &Roadmap{Phases: []Phase{{Name: "P0", Features: []Feature{{Name: "user"}, {Name: "post"}}}}}

	cases := []struct{ name, features, want string }{
		{"unknown", `{"name":"ghost",` + goodScores + `,"summary":"s"}`, "unknown feature"},
		{"badscore", `{"name":"user","scores":{"acceptanceCriteria":0,"userValue":9,"definitionClarity":9,"dependencies":9,"effortEstimation":9,"overall":9},"summary":"s"}`, "invalid scores"},
		{"nosummary", `{"name":"user",` + goodScores + `,"summary":"  "}`, "summary is required"},
		{"missing", `{"name":"user",` + goodScores + `,"summary":"s"}`, "missing feature"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if err := os.WriteFile(RoadmapQualityFile, []byte(qualityReport(c.features)), 0644); err != nil {
				t.Fatal(err)
			}
			err := ValidateQuality(r)
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("want %q, got %v", c.want, err)
			}
		})
	}
}
