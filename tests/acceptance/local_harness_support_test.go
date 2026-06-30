package acceptance_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/setup"
)

// AC#8 (HERMETIC, no network): an [orchestration.local] ollama block resolves the
// declared model to limited → strict and wires a managed opencode.json provider at
// the configured baseURL, while Claude/Aider managed files stay untouched. No real
// network call is made — config is shape-only and the provider is wired at the file
// seam (the runner's job to reach the endpoint).
func TestLocalHarnessSupport_EndToEndHermetic(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck
	os.Chdir(dir)        //nolint:errcheck

	toml := "[orchestration.local]\nprovider = \"ollama\"\nendpoint = \"http://localhost:11434/v1\"\nmodel = \"qwen2.5-coder\"\n"
	os.WriteFile("centinela.toml", []byte(toml), 0644) //nolint:errcheck

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if got := config.DriverModelFrom("", cfg); got != "qwen2.5-coder" {
		t.Fatalf("driver model = %q, want qwen2.5-coder", got)
	}
	if p, ok := config.DefaultProfileForModel("qwen2.5-coder", cfg); !ok || p != config.ProfileStrict {
		t.Fatalf("limited → strict profile = (%q,%v), want (strict,true)", p, ok)
	}

	lc, _ := config.LocalProviderConfig(cfg)
	lp := &setup.LocalProvider{Provider: lc.Provider, Endpoint: lc.Endpoint, Model: lc.Model, APIKeyEnv: lc.APIKeyEnv}
	plan, err := setup.BuildSyncPlanWithLocal("opencode", lp)
	if err != nil {
		t.Fatalf("plan: %v", err)
	}
	if err := setup.ApplySync(plan); err != nil {
		t.Fatalf("apply: %v", err)
	}

	data, err := os.ReadFile("opencode.json")
	if err != nil {
		t.Fatalf("read opencode.json: %v", err)
	}
	var top map[string]json.RawMessage
	json.Unmarshal(data, &top) //nolint:errcheck
	var providers map[string]map[string]any
	json.Unmarshal(top["provider"], &providers) //nolint:errcheck
	opts, _ := providers["ollama"]["options"].(map[string]any)
	if opts == nil || opts["baseURL"] != "http://localhost:11434/v1" {
		t.Fatalf("managed provider baseURL wrong: %#v", providers["ollama"])
	}
	if _, err := os.Stat(".claude/settings.json"); !os.IsNotExist(err) {
		t.Fatal(".claude/settings.json must be untouched")
	}
	if _, err := os.Stat(".aider.conf.yml"); !os.IsNotExist(err) {
		t.Fatal(".aider.conf.yml must be untouched")
	}
}
