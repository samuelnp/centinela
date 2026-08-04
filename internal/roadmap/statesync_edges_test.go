package roadmap

import (
	"os"
	"strings"
	"testing"
)

// A missing committer is a wiring bug, not a policy choice: it must warn.
func TestSyncWithoutACommitterWarns(t *testing.T) {
	syncRepo(t, syncBody)
	rep := Sync(SyncOptions{Verb: "defer", Subject: "x", Commit: true})
	if !rep.Warn || !strings.Contains(rep.Reason, "no committer") {
		t.Fatalf("report = %+v", rep)
	}
	if !rep.Regenerated {
		t.Fatal("regeneration must still have happened")
	}
}

// A ROADMAP.md that cannot be written surfaces as a warning, never a crash.
func TestSyncSurfacesAMarkdownWriteFailure(t *testing.T) {
	syncRepo(t, syncBody)
	if err := os.Mkdir("ROADMAP.md", 0o755); err != nil {
		t.Fatal(err)
	}
	rep := Sync(SyncOptions{Verb: "defer", Subject: "x", Commit: true, C: &fakeCommitter{}})
	if !rep.Warn || !strings.Contains(rep.Reason, "regeneration failed") {
		t.Fatalf("report = %+v", rep)
	}
}

func TestWriteRoadmapJSONReplacesTheFileAtomically(t *testing.T) {
	syncRepo(t, syncBody)
	merged := []byte(`{"phases":[{"name":"P","features":[]}]}`)
	if err := WriteRoadmapJSON(merged); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(RoadmapFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(merged) {
		t.Fatalf("roadmap.json = %q", got)
	}
	if _, err := Load(); err != nil {
		t.Fatalf("the written document must still load: %v", err)
	}
}

// F1: the report must carry a VERIFIED read-back, not an assumption.
func TestSyncVerifiesTheStateIsOnDisk(t *testing.T) {
	syncRepo(t, syncBody)
	rep := Sync(SyncOptions{Verb: "defer", Subject: "x", Commit: false})
	if !rep.InWorkingTree {
		t.Fatal("a completed regeneration must verify as in the working tree")
	}
	if !StateInSync() {
		t.Fatal("StateInSync must agree")
	}
	// Simulate a concurrent process replacing roadmap.json after this mutation:
	// the read-back must now refuse to vouch for the tree.
	if err := os.WriteFile(RoadmapFile,
		[]byte(`{"phases":[{"name":"Other","features":[]}]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if StateInSync() {
		t.Fatal("a roadmap.json rewritten underneath must not report as in sync")
	}
}

// An unreadable or absent ROADMAP.md can never be vouched for either.
func TestStateInSyncFailsClosed(t *testing.T) {
	syncRepo(t, syncBody)
	if StateInSync() {
		t.Fatal("no ROADMAP.md at all must not report as in sync")
	}
	if _, err := RegenerateMarkdown(); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(RoadmapFile, []byte("{broken"), 0o644); err != nil {
		t.Fatal(err)
	}
	if StateInSync() {
		t.Fatal("an unparseable roadmap.json must not report as in sync")
	}
}
