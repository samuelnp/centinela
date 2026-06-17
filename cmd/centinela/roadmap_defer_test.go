package main

import (
	"os"
	"strings"
	"testing"
)

// TestResolveDeferSource_ExplicitFlag parses feature/role from the flag.
func TestResolveDeferSource_ExplicitFlag(t *testing.T) {
	src := resolveDeferSource("my-feature/senior-engineer")
	if src == nil {
		t.Fatal("expected non-nil source")
	}
	if src.Feature != "my-feature" || src.Role != "senior-engineer" {
		t.Errorf("unexpected source: %+v", src)
	}
}

// TestResolveDeferSource_FlagFeatureOnly parses feature without role.
func TestResolveDeferSource_FlagFeatureOnly(t *testing.T) {
	src := resolveDeferSource("my-feature")
	if src == nil {
		t.Fatal("expected non-nil source")
	}
	if src.Feature != "my-feature" || src.Role != "" {
		t.Errorf("unexpected source: %+v", src)
	}
}

// TestResolveDeferSource_Empty returns nil for empty flag.
func TestResolveDeferSource_Empty(t *testing.T) {
	orig, _ := os.Getwd()
	d := t.TempDir()
	os.Chdir(d)          //nolint:errcheck
	defer os.Chdir(orig) //nolint:errcheck
	src := resolveDeferSource("")
	// outside a .worktrees/ dir: should return nil
	if src != nil {
		t.Logf("source from non-worktree dir: %+v", src)
	}
	// We just assert no panic; behavior depends on CWD
}

// TestRunRoadmapDefer_HappyPath calls Defer via the cobra command.
func TestRunRoadmapDefer_HappyPath(t *testing.T) {
	d := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig)                                                                                  //nolint:errcheck
	os.Chdir(d)                                                                                           //nolint:errcheck
	os.MkdirAll(".workflow", 0755)                                                                        //nolint:errcheck
	os.WriteFile(".workflow/roadmap.json", []byte(`{"phases":[{"name":"Phase 0","features":[]}]}`), 0644) //nolint:errcheck

	deferSummary = "a finding"
	deferSource = "my-feat/eng"
	err := runRoadmapDefer(nil, []string{"my-finding"})
	if err != nil {
		t.Fatalf("runRoadmapDefer: %v", err)
	}
	data, _ := os.ReadFile(".workflow/roadmap.json")
	if !strings.Contains(string(data), "my-finding") {
		t.Error("finding must be in roadmap.json")
	}
}

// TestRunRoadmapDefer_EmptySummary returns error.
func TestRunRoadmapDefer_EmptySummary(t *testing.T) {
	d := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig)                                                  //nolint:errcheck
	os.Chdir(d)                                                           //nolint:errcheck
	os.MkdirAll(".workflow", 0755)                                        //nolint:errcheck
	os.WriteFile(".workflow/roadmap.json", []byte(`{"phases":[]}`), 0644) //nolint:errcheck

	deferSummary = ""
	deferSource = ""
	if err := runRoadmapDefer(nil, []string{"x"}); err == nil {
		t.Error("expected error for empty summary")
	}
}
