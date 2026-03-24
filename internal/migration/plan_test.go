package migration

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuildPlanDetectsLegacyAndApplyWrites(t *testing.T) {
	d := t.TempDir()
	os.MkdirAll(filepath.Join(d, "docs", "architecture"), 0755)                         //nolint:errcheck
	os.WriteFile(filepath.Join(d, "PROJECT.md.template"), []byte("# template\n"), 0644) //nolint:errcheck

	plan, err := BuildPlan(d)
	if err != nil {
		t.Fatal(err)
	}
	if !plan.HasChanges() {
		t.Fatal("expected migration plan changes")
	}
	if err := Apply(d, plan); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(d, "CLAUDE.md"))
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := ParseHeader(string(b)); !ok {
		t.Fatal("expected CLAUDE.md to include migration header")
	}
}
