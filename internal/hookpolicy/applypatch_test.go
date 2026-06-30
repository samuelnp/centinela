package hookpolicy

import (
	"reflect"
	"testing"
)

func TestExtractApplyPatchPaths_AllVerbs(t *testing.T) {
	env := "*** Begin Patch\n" +
		"*** Add File: internal/a.go\n" +
		"*** Update File: internal/b.go\n" +
		"*** Delete File: internal/c.go\n" +
		"*** Move to: internal/d.go\n" +
		"*** End Patch"
	got := ExtractApplyPatchPaths(env)
	want := []string{"internal/a.go", "internal/b.go", "internal/c.go", "internal/d.go"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestExtractApplyPatchPaths_MultiFileSingleVerb(t *testing.T) {
	env := "*** Add File: README.md\n*** Add File: src/x.ts\n"
	got := ExtractApplyPatchPaths(env)
	if !reflect.DeepEqual(got, []string{"README.md", "src/x.ts"}) {
		t.Fatalf("multi-file = %v", got)
	}
}

func TestExtractApplyPatchPaths_None(t *testing.T) {
	if got := ExtractApplyPatchPaths("just some text\nno patch verbs"); got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
}

func TestExtractApplyPatchPaths_WhitespaceTrimmed(t *testing.T) {
	env := "   *** Add File:    internal/spaced.go   \n"
	got := ExtractApplyPatchPaths(env)
	if !reflect.DeepEqual(got, []string{"internal/spaced.go"}) {
		t.Fatalf("trim failed: %q", got)
	}
}

func TestExtractApplyPatchPaths_EmptyPathSkipped(t *testing.T) {
	if got := ExtractApplyPatchPaths("*** Add File: \n"); got != nil {
		t.Fatalf("empty path should be skipped, got %v", got)
	}
}
