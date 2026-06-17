package gates

import (
	"os"
	"strings"
	"testing"
)

func TestCheckRoadmapDriftLoadError(t *testing.T) {
	t.Chdir(t.TempDir())
	r := checkRoadmapDrift(driftCfg("warn"), nil)
	if r.Status != Fail || !strings.Contains(r.Message, "load") {
		t.Fatalf("missing roadmap.json must Fail, got %v %q", r.Status, r.Message)
	}
}

// A non-missing read error (ROADMAP.md is a directory) Fails, not a panic.
func TestCheckRoadmapDriftReadError(t *testing.T) {
	seedDrift(t, nil)
	if err := os.Mkdir("ROADMAP.md", 0o755); err != nil {
		t.Fatal(err)
	}
	r := checkRoadmapDrift(driftCfg("warn"), nil)
	if r.Status != Fail || !strings.Contains(r.Message, "cannot read") {
		t.Fatalf("read error must Fail with 'cannot read', got %v %q", r.Status, r.Message)
	}
}

func TestFirstDifferingLine(t *testing.T) {
	cases := []struct {
		name      string
		want, got string
		expect    int
	}{
		{"identical", "a\nb\nc\n", "a\nb\nc\n", 0},
		{"first line", "x\nb\n", "a\nb\n", 1},
		{"middle line", "a\nb\nc\n", "a\nZ\nc\n", 2},
		{"missing trailing line", "a\nb\nc\n", "a\nb\n", 3},
		{"extra trailing line", "a\nb\n", "a\nb\nc\n", 3},
		{"both empty", "", "", 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := firstDifferingLine([]byte(c.want), []byte(c.got)); got != c.expect {
				t.Fatalf("firstDifferingLine(%q,%q)=%d want %d", c.want, c.got, got, c.expect)
			}
		})
	}
}
