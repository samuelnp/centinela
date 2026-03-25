package autostart

import "testing"

func TestExtractPromptAndIntent(t *testing.T) {
	raw := []byte(`{"prompt":"please add github release automation"}`)
	if got := ExtractPrompt(raw); got != "please add github release automation" {
		t.Fatalf("unexpected prompt: %q", got)
	}
	if !ShouldStart("I want to add release notes") {
		t.Fatal("expected start intent")
	}
	if ShouldStart("step plan is done shall I advance?") {
		t.Fatal("review prompt should not trigger start")
	}
}

func TestDeriveFeature(t *testing.T) {
	name := DeriveFeature("Please add Github workflow release automation for windows")
	if name != "github-workflow-release-automation-for-windows" {
		t.Fatalf("unexpected feature name: %q", name)
	}
	if got := DeriveFeature("new feature: 123 !!!"); got == "" {
		t.Fatal("expected fallback feature name")
	}
}
