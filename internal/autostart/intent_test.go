package autostart

import "testing"

func TestExtractPromptAndIntent(t *testing.T) {
	raw := []byte(`{"prompt":"please add github release automation"}`)
	if got := ExtractPrompt(raw); got != "please add github release automation" {
		t.Fatalf("unexpected prompt: %q", got)
	}
	if got := ExtractPrompt([]byte(`{"input":{"text":"extend release pipeline"}}`)); got != "extend release pipeline" {
		t.Fatalf("unexpected nested prompt: %q", got)
	}
	if got := ExtractPrompt([]byte("plain text request")); got != "plain text request" {
		t.Fatalf("unexpected plain prompt: %q", got)
	}
	if got := ExtractPrompt([]byte(`{"message":"new feature parser"}`)); got != "new feature parser" {
		t.Fatalf("unexpected message prompt: %q", got)
	}
	if got := ExtractPrompt([]byte(`{"input":"extend cli"}`)); got != "extend cli" {
		t.Fatalf("unexpected input prompt: %q", got)
	}
	if got := ExtractPrompt([]byte(`{"prompt":`)); got != `{"prompt":` {
		t.Fatalf("unexpected invalid-json fallback: %q", got)
	}
	if !ShouldStart("I want to add release notes") {
		t.Fatal("expected start intent")
	}
	if !ShouldStart("new feature for reports") {
		t.Fatal("expected new feature intent")
	}
	if ShouldStart("") {
		t.Fatal("empty prompt should not start")
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
	if got := DeriveFeature("123 release diagnostics alpha"); got != "feature-123-release-diagnostics-alpha" {
		t.Fatalf("expected numeric prefix guard, got %q", got)
	}
	if got := DeriveFeature("new feature: 123 !!!"); got == "" {
		t.Fatal("expected fallback feature name")
	}
}
