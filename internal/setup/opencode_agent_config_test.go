package setup

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestOpenCodeAgentsCarryVerifierAndLegacyRole(t *testing.T) {
	gk, ok := centinelaOpenCodeAgents["gatekeeper"]
	if !ok {
		t.Fatal("gatekeeper agent entry missing")
	}
	for _, want := range []string{"refutes", "adversarial"} {
		if !strings.Contains(strings.ToLower(gk["description"]), want) {
			t.Fatalf("gatekeeper description missing %q: %s", want, gk["description"])
		}
	}
	if !strings.Contains(gk["prompt"], "centinela artifact stamp") {
		t.Fatalf("gatekeeper prompt must mandate the stamp: %s", gk["prompt"])
	}
	// Legacy workflows and adapt_opencode_support still enumerate this role.
	if _, ok := centinelaOpenCodeAgents["validation-specialist"]; !ok {
		t.Fatal("validation-specialist must be retained for legacy workflows")
	}
}

func TestMergeOpenCodeAgentsAllowsVerifierTask(t *testing.T) {
	raw := map[string]json.RawMessage{}
	if !mergeOpenCodeAgents(raw) {
		t.Fatal("merge into an empty config must report a change")
	}
	agents := map[string]json.RawMessage{}
	if err := json.Unmarshal(raw["agent"], &agents); err != nil {
		t.Fatal(err)
	}
	if _, ok := agents["gatekeeper"]; !ok {
		t.Fatalf("gatekeeper agent not emitted: %v", agents)
	}
	build := map[string]json.RawMessage{}
	_ = json.Unmarshal(agents["build"], &build)
	permission := map[string]json.RawMessage{}
	_ = json.Unmarshal(build["permission"], &permission)
	task := map[string]string{}
	_ = json.Unmarshal(permission["task"], &task)
	if task["gatekeeper"] != "allow" {
		t.Fatalf("gatekeeper task permission = %q, want allow", task["gatekeeper"])
	}
}

func TestMergeOpenCodeAgentsIsIdempotent(t *testing.T) {
	raw := map[string]json.RawMessage{}
	mergeOpenCodeAgents(raw)
	if mergeOpenCodeAgents(raw) {
		t.Fatal("a second merge must report no change (managed-sync would report perpetual drift)")
	}
}
