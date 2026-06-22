package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// runRecon invokes runReconstruct with buffered output and restores the flags.
func runRecon(t *testing.T, in, out string, asJSON bool) (string, error) {
	t.Helper()
	oi, oo, oj := reconIn, reconOut, reconJSON
	t.Cleanup(func() { reconIn, reconOut, reconJSON = oi, oo, oj })
	reconIn, reconOut, reconJSON = in, out, asJSON
	var buf bytes.Buffer
	c := &cobra.Command{}
	c.SetOut(&buf)
	err := runReconstruct(c, nil)
	return buf.String(), err
}

func TestReconstruct_WritesCorpusSummary(t *testing.T) {
	in := writeInventory(t, ntierInventory)
	out := filepath.Join(t.TempDir(), "review")
	stdout, err := runRecon(t, in, out, false)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"targets selected:", "files written:", "TODO confirm markers:"} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("summary missing %q:\n%s", want, stdout)
		}
	}
	if _, err := os.Stat(filepath.Join(out, "specs")); err != nil {
		t.Fatalf("specs review dir not written: %v", err)
	}
	if entries, _ := os.ReadDir(filepath.Join(out, "features")); len(entries) == 0 {
		t.Fatal("brief stubs not written")
	}
}

func TestReconstruct_JSON(t *testing.T) {
	out := filepath.Join(t.TempDir(), "review")
	stdout, err := runRecon(t, writeInventory(t, ntierInventory), out, true)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"targets"`, `"todoCount"`, `"written"`} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("json missing %q:\n%s", want, stdout)
		}
	}
}
