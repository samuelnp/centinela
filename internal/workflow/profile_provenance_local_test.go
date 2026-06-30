package workflow

import (
	"testing"

	"github.com/samuelnp/centinela/internal/config"
)

func cfgLocalModel(model string) *config.Config {
	c := &config.Config{}
	c.Orchestration.Local = config.LocalConfig{Provider: "ollama", Endpoint: "http://x/v1", Model: model}
	return c
}

// ProfileProvenance emits the local-default note when an unmapped local model
// drives and no explicit profile is set; omits it when an explicit per-feature
// profile wins or when the model carries an explicit capability mapping.
func TestProfileProvenanceLocalDefault(t *testing.T) {
	local := cfgLocalModel("qwen2.5-coder")
	p, note := ProfileProvenance(&Workflow{DriverModel: "qwen2.5-coder"}, local)
	if p != config.ProfileStrict || note != "local default: qwen2.5-coder → limited → strict" {
		t.Fatalf("local default: got (%q,%q)", p, note)
	}

	p2, note2 := ProfileProvenance(&Workflow{DriverModel: "qwen2.5-coder", EnforcementProfile: config.ProfileOutcome}, local)
	if p2 != config.ProfileOutcome || note2 != "--profile" {
		t.Fatalf("explicit --profile must win: got (%q,%q)", p2, note2)
	}

	mapped := cfgLocalModel("qwen2.5-coder")
	mapped.Orchestration.Capabilities = map[string]string{"qwen2.5-coder": "capable"}
	_, note3 := ProfileProvenance(&Workflow{DriverModel: "qwen2.5-coder"}, mapped)
	if note3 != "driver: qwen2.5-coder → capable" {
		t.Fatalf("explicitly mapped model must use the driver note: got %q", note3)
	}
}
