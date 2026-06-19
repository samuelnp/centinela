package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

const ntierInventory = `{"schemaVersion":1,"primaryLanguage":"Go",
"manifests":[{"kind":"go-mod","path":"go.mod","build":"go build","test":"go test"}],
"packages":["internal/handler","internal/service","internal/repository"],
"graph":{"kind":"go-packages","edges":[]}}`

// writeInventory writes an analysis.json into a fresh temp dir and returns its path.
func writeInventory(t *testing.T, body string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "analysis.json")
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

// runSynth invokes runSynthesize with buffered output and restores the flags.
func runSynth(t *testing.T, in, out string, asJSON bool) (string, error) {
	t.Helper()
	oi, oo, oj := synthIn, synthOut, synthJSON
	t.Cleanup(func() { synthIn, synthOut, synthJSON = oi, oo, oj })
	synthIn, synthOut, synthJSON = in, out, asJSON
	var buf bytes.Buffer
	c := &cobra.Command{}
	c.SetOut(&buf)
	err := runSynthesize(c, nil)
	return buf.String(), err
}

func TestSynthesize_WritesDraft(t *testing.T) {
	in := writeInventory(t, ntierInventory)
	out := filepath.Join(t.TempDir(), "PROJECT.md")
	stdout, err := runSynth(t, in, out, false)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout, "n-tier") || !strings.Contains(stdout, "wrote "+out) {
		t.Fatalf("summary wrong:\n%s", stdout)
	}
	if b, _ := os.ReadFile(out); !strings.Contains(string(b), "**Archetype:** n-tier") {
		t.Fatalf("draft not written: %q", b)
	}
}

func TestSynthesize_JSON(t *testing.T) {
	stdout, err := runSynth(t, writeInventory(t, ntierInventory), "", true)
	if err != nil || !strings.Contains(stdout, `"Best": "n-tier"`) {
		t.Fatalf("json output wrong: %q err=%v", stdout, err)
	}
}
