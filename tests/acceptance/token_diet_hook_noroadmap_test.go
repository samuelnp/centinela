// Acceptance: specs/token-diet.feature
package acceptance_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// Scenario: An unwritable state directory never breaks the host session
func TestTD_UnwritableStateDirectoryNeverBreaksHostSession(t *testing.T) {
	if runtime.GOOS == "windows" || os.Geteuid() == 0 {
		t.Skip("permission bits are not enforced for this test's identity")
	}
	bin := tdBuildBin(t)
	dir := tdRepo(t)
	tdWriteRoadmap(t, dir, tdRoadmap1)
	if err := os.Chmod(filepath.Join(dir, ".workflow"), 0o500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(filepath.Join(dir, ".workflow"), 0o755) }) //nolint:errcheck

	out, code := tdHookContext(t, bin, dir, tdSessionPayload("s-1"))
	if code != 0 {
		t.Fatalf("hook must exit 0 even with an unwritable .workflow/: %s", out)
	}
	mustContain(t, out, tdRoadmapLine)
	mustNotContain(t, out, "error")
}

// Scenario: A missing or invalid roadmap prints nothing and writes no state
func TestTD_MissingOrInvalidRoadmapPrintsNothingWritesNoState(t *testing.T) {
	bin := tdBuildBin(t)
	for _, mutate := range []func(dir string){
		func(dir string) {}, // no roadmap.json at all
		func(dir string) { mustWrite(t, filepath.Join(dir, ".workflow", "roadmap.json"), "{not valid json") },
	} {
		dir := tdRepo(t)
		mutate(dir)
		out, code := tdHookContext(t, bin, dir, tdSessionPayload("s-1"))
		if code != 0 {
			t.Fatalf("hook must exit 0: %s", out)
		}
		mustNotContain(t, out, tdRoadmapLine)
		if _, err := os.Stat(tdDigestPath(dir)); !os.IsNotExist(err) {
			t.Fatalf("no digest state should be written when the roadmap is absent/invalid")
		}
	}
}
