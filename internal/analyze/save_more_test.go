package analyze

import (
	"os"
	"testing"
)

func TestSave_NoDirectoryComponentSkipsMkdir(t *testing.T) {
	// A bare filename has dir "." so Save skips MkdirAll entirely and writes in
	// the cwd. Run inside a temp dir so the artifact is cleaned up automatically.
	dir := t.TempDir()
	orig, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(orig) })
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	if err := Save("analysis.json", sampleInventory()); err != nil {
		t.Fatalf("Save of a bare filename must succeed: %v", err)
	}
	if _, err := os.Stat("analysis.json"); err != nil {
		t.Fatalf("expected file written in cwd: %v", err)
	}
}
