package migration

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuildPlanNoChangesAfterApply(t *testing.T) {
	d := t.TempDir()
	plan, err := BuildPlan(d)
	if err != nil {
		t.Fatal(err)
	}
	if err := Apply(d, plan); err != nil {
		t.Fatal(err)
	}
	plan2, err := BuildPlan(d)
	if err != nil {
		t.Fatal(err)
	}
	if plan2.HasChanges() {
		t.Fatal("expected no changes after apply")
	}
}

func TestBuildPlanUpdatesWrongTemplateHeader(t *testing.T) {
	d := t.TempDir()
	bad := "<!-- centinela:doc-version=1 template=wrong.md -->\n# x\n"
	os.WriteFile(filepath.Join(d, "CLAUDE.md"), []byte(bad), 0644) //nolint:errcheck
	plan, err := BuildPlan(d)
	if err != nil {
		t.Fatal(err)
	}
	if !plan.HasChanges() {
		t.Fatal("expected update for wrong template header")
	}
}
