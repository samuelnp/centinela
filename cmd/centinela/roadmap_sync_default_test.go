package main

import "testing"

// An absent centinela.toml is a zero-config project: auto-commit stays on.
func TestRoadmapAutoCommitDefaultsOn(t *testing.T) {
	t.Chdir(t.TempDir())
	if !roadmapAutoCommitEnabled() {
		t.Fatal("auto-commit must default to enabled")
	}
}
