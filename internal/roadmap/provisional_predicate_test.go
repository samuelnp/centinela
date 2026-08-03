package roadmap

import (
	"os"
	"testing"
)

// TestArtifactIsProvisional is the predicate the STRICT setup cascade shares
// with ValidateAnalysis/ValidateQuality. Crucially, an absent or unreadable file
// is NOT provisional: turning "missing" into "seeded" would make the cascade
// report the wrong reason for a tree that simply has no artifacts yet.
func TestArtifactIsProvisional(t *testing.T) {
	t.Chdir(t.TempDir())
	os.MkdirAll(".workflow", 0755) //nolint:errcheck
	cases := []struct {
		name, body string
		want       bool
	}{
		{"seeded", `{"role":"x","provisional":true,"features":[]}`, true},
		{"explicitly not provisional", `{"role":"x","provisional":false,"features":[]}`, false},
		{"evaluated (key absent)", `{"role":"x","features":[]}`, false},
		{"malformed json", `{not json`, false},
		{"empty file", ``, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			os.WriteFile(RoadmapAnalysisFile, []byte(tc.body), 0644) //nolint:errcheck
			if got := ArtifactIsProvisional(RoadmapAnalysisFile); got != tc.want {
				t.Fatalf("ArtifactIsProvisional = %v, want %v", got, tc.want)
			}
		})
	}
	os.Remove(RoadmapAnalysisFile) //nolint:errcheck
	if ArtifactIsProvisional(RoadmapAnalysisFile) {
		t.Fatal("an ABSENT artifact must not read as provisional")
	}
	// The mark the seeder actually writes must be detected by this predicate.
	if err := SeedArtifactsIfAbsent(); err != nil {
		t.Fatalf("seed: %v", err)
	}
	for _, p := range []string{RoadmapAnalysisFile, RoadmapQualityFile} {
		if !ArtifactIsProvisional(p) {
			t.Fatalf("%s: the seeder's own mark must be detected", p)
		}
	}
}
