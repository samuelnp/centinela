package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// runBrown invokes runRoadmapBrownfield with buffered output, restoring flags.
func runBrown(t *testing.T, in, out string, asJSON bool, goals []string) (string, error) {
	t.Helper()
	oi, oo, oj, og := brownIn, brownOut, brownJSON, brownGoals
	t.Cleanup(func() { brownIn, brownOut, brownJSON, brownGoals = oi, oo, oj, og })
	brownIn, brownOut, brownJSON, brownGoals = in, out, asJSON, goals
	var buf bytes.Buffer
	c := &cobra.Command{}
	c.SetOut(&buf)
	err := runRoadmapBrownfield(c, nil)
	return buf.String(), err
}

func TestRoadmapBrownfield_WritesDraftLeavesCanonicalUntouched(t *testing.T) {
	dir := t.TempDir()
	in := writeInventory(t, ntierInventory)
	canonical := filepath.Join(dir, "roadmap.json")
	const curated = `{"phases":[{"name":"Phase 1","features":[{"name":"hand-authored"}]}]}`
	if err := os.WriteFile(canonical, []byte(curated), 0o644); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(dir, "roadmap.brownfield.json")
	stdout, err := runBrown(t, in, out, false, []string{"Add OAuth login"})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"baseline entries:", "gaps:", "draft written:"} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("summary missing %q:\n%s", want, stdout)
		}
	}
	if _, err := os.Stat(out); err != nil {
		t.Fatalf("draft not written: %v", err)
	}
	after, _ := os.ReadFile(canonical)
	if string(after) != curated {
		t.Fatal("canonical roadmap.json must be left byte-for-byte unchanged")
	}
}

func TestRoadmapBrownfield_JSON(t *testing.T) {
	out := filepath.Join(t.TempDir(), "draft.json")
	stdout, err := runBrown(t, writeInventory(t, ntierInventory), out, true, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"BaselineCount"`, `"GapCount"`, `"DraftPath"`} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("json missing %q:\n%s", want, stdout)
		}
	}
}
