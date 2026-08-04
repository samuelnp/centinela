package roadmap

import (
	"os"
	"strings"
	"testing"
)

// fakeCommitter records what Sync asked for and answers with a fixed result.
type fakeCommitter struct {
	calls int
	msg   string
	paths []string
	err   error
}

func (f *fakeCommitter) Commit(msg string, paths []string) error {
	f.calls++
	f.msg, f.paths = msg, paths
	return f.err
}

// syncRepo chdirs into a temp dir holding roadmap.json, so Sync's cwd-relative
// reads and writes stay inside the test.
func syncRepo(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(dir+"/.workflow", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dir+"/"+RoadmapFile, []byte(body), 0644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)
	return dir
}

const syncBody = `{"phases":[{"name":"Phase 1","features":[{"name":"a","dependsOn":[]}]}]}`

func TestSyncRegeneratesAndCommits(t *testing.T) {
	syncRepo(t, syncBody)
	c := &fakeCommitter{}
	rep := Sync(SyncOptions{Verb: "defer", Subject: "flaky-thing", Commit: true, C: c})
	if !rep.Regenerated || !rep.Committed || rep.Warn {
		t.Fatalf("report = %+v", rep)
	}
	md, err := os.ReadFile("ROADMAP.md")
	if err != nil || !strings.Contains(string(md), "Phase 1") {
		t.Fatalf("ROADMAP.md not regenerated: %v %q", err, md)
	}
	if c.calls != 1 || c.msg != "chore(roadmap): defer flaky-thing" {
		t.Fatalf("commit call = %d %q", c.calls, c.msg)
	}
	if strings.Join(c.paths, ",") != ".workflow/roadmap.json,ROADMAP.md" {
		t.Fatalf("pathspec = %v", c.paths)
	}
}

func TestSyncForwardsExtraPaths(t *testing.T) {
	syncRepo(t, syncBody)
	c := &fakeCommitter{}
	Sync(SyncOptions{Verb: "promote", Subject: "x", ExtraPaths: PromoteArtifactPaths(), Commit: true, C: c})
	got := strings.Join(c.paths, ",")
	for _, want := range append([]string{".workflow/roadmap.json", "ROADMAP.md"}, PromoteArtifactPaths()...) {
		if !strings.Contains(got, want) {
			t.Fatalf("pathspec %q is missing %q", got, want)
		}
	}
}
