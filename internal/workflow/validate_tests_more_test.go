package workflow

import (
	"os"
	"testing"
)

func TestHasAcceptanceTestsSuffixAndAny(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	os.MkdirAll("tests/acceptance", 0755)                                                                    //nolint:errcheck
	os.WriteFile("tests/acceptance/a.steps.ts", []byte("Given('x', () => {})\nThen('y', () => {})\n"), 0644) //nolint:errcheck
	if !hasAcceptanceTests(".steps.ts") {
		t.Fatal("expected acceptance suffix match")
	}
	if !hasAcceptanceTests("") {
		t.Fatal("expected acceptance any-file match")
	}
	if hasAnyFile("tests/missing") {
		t.Fatal("missing dir should report no files")
	}
}
