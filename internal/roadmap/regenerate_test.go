package roadmap

import (
	"os"
	"strings"
	"testing"
)

func TestRegenerateMarkdownIsIdempotent(t *testing.T) {
	syncRepo(t, syncBody)
	changed, err := RegenerateMarkdown()
	if err != nil || !changed {
		t.Fatalf("first render: changed=%v err=%v", changed, err)
	}
	first, err := os.ReadFile("ROADMAP.md")
	if err != nil {
		t.Fatal(err)
	}
	changed, err = RegenerateMarkdown()
	if err != nil {
		t.Fatal(err)
	}
	if changed {
		t.Fatal("a second render of unchanged state must report no change (no empty commit)")
	}
	second, _ := os.ReadFile("ROADMAP.md") //nolint:errcheck
	if string(first) != string(second) {
		t.Fatal("rendering must be byte-stable")
	}
}

func TestRegenerateMarkdownMatchesTheDriftOracle(t *testing.T) {
	syncRepo(t, syncBody)
	if _, err := RegenerateMarkdown(); err != nil {
		t.Fatal(err)
	}
	r, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	onDisk, err := os.ReadFile("ROADMAP.md")
	if err != nil {
		t.Fatal(err)
	}
	if string(onDisk) != string(RenderMarkdown(r)) {
		t.Fatal("ROADMAP.md must byte-match the renderer the drift gate compares against")
	}
}

func TestRegenerateMarkdownSurfacesAnUnreadableRoadmap(t *testing.T) {
	syncRepo(t, "{not json")
	if _, err := RegenerateMarkdown(); err == nil {
		t.Fatal("an unparseable roadmap.json must surface, not silently render nothing")
	}
	rep := Sync(SyncOptions{Verb: "defer", Subject: "x", Commit: true, C: &fakeCommitter{}})
	if !rep.Warn || !strings.Contains(rep.Reason, "regeneration failed") {
		t.Fatalf("report = %+v", rep)
	}
}

// R7: the declared pathspec must name exactly the artifacts promote rewrites.
func TestPromoteArtifactPathsMatchesWhatPromoteWrites(t *testing.T) {
	got := strings.Join(PromoteArtifactPaths(), ",")
	want := strings.Join([]string{
		RoadmapAnalysisFile, RoadmapQualityFile,
		RoadmapAnalysisMarkdown, RoadmapQualityMarkdown,
	}, ",")
	if got != want {
		t.Fatalf("PromoteArtifactPaths = %q, want %q", got, want)
	}
	for _, p := range PromoteArtifactPaths() {
		if !strings.HasPrefix(p, ".workflow/roadmap-") {
			t.Fatalf("%q is outside the roadmap-state family", p)
		}
	}
}
