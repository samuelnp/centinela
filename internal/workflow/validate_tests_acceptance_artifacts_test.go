package workflow

import (
	"os"
	"testing"
)

func TestHasExecutableAcceptanceContentBranches(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	os.MkdirAll("tests/acceptance", 0755)                                                //nolint:errcheck
	os.WriteFile("tests/acceptance/comments.go", []byte("/* block */\n// line\n"), 0644) //nolint:errcheck
	if hasExecutableAcceptanceContent("tests/acceptance/comments.go") {
		t.Fatal("expected comment-only acceptance content to be rejected")
	}
	os.WriteFile("tests/acceptance/todo.go", []byte("func TestX() { /* TODO */ }"), 0644) //nolint:errcheck
	if hasExecutableAcceptanceContent("tests/acceptance/todo.go") {
		t.Fatal("expected placeholder TODO acceptance content to be rejected")
	}
	if hasExecutableAcceptanceContent("tests/acceptance/missing.go") {
		t.Fatal("expected missing acceptance file to be rejected")
	}
}

func TestLooksLikeExecutableTestFalse(t *testing.T) {
	if looksLikeExecutableTest("plain narrative without executable checks") {
		t.Fatal("expected non-test text to be rejected")
	}
}
