package workflow

import (
	"os"
	"testing"
)

func TestHasEdgeCaseReport(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	if hasEdgeCaseReport("f") {
		t.Fatal("report should be absent")
	}
	os.MkdirAll(".workflow", 0755)                                //nolint:errcheck
	os.WriteFile(".workflow/f-edge-cases.md", []byte("ok"), 0644) //nolint:errcheck
	if !hasEdgeCaseReport("f") {
		t.Fatal("report should be detected")
	}
	if hasEdgeCaseReport("") {
		t.Fatal("empty feature should not resolve report")
	}
}
