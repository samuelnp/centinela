package setup

import (
	"encoding/json"
	"strings"
	"testing"
)

func plannerAgentsOf(t *testing.T, raw map[string]json.RawMessage) map[string]json.RawMessage {
	t.Helper()
	out := map[string]json.RawMessage{}
	if err := json.Unmarshal(raw["agent"], &out); err != nil {
		t.Fatalf("unmarshal agents: %v", err)
	}
	return out
}

// D10: the emitted set carries planner and no longer carries the retired pair.
func TestOpenCodeAgents_EmitsPlannerNotLegacyPair(t *testing.T) {
	planner, ok := centinelaOpenCodeAgents["planner"]
	if !ok {
		t.Fatal("planner agent must be emitted")
	}
	for _, retired := range []string{"big-thinker", "feature-specialist"} {
		if _, ok := centinelaOpenCodeAgents[retired]; ok {
			t.Errorf("retired agent %q must no longer be emitted", retired)
		}
	}
	for _, want := range []string{"Lens 1", "Lens 2", "strategy", "spec"} {
		if !strings.Contains(planner["prompt"], want) {
			t.Errorf("planner prompt missing %q", want)
		}
	}
}

func TestMergeOpenCodeAgents_AddsPlannerAndTaskPermission(t *testing.T) {
	raw := map[string]json.RawMessage{}
	if !mergeOpenCodeAgents(raw) {
		t.Fatal("merge into an empty config must report a change")
	}
	agents := plannerAgentsOf(t, raw)
	if _, ok := agents["planner"]; !ok {
		t.Fatal("planner agent not written")
	}
	build := map[string]json.RawMessage{}
	_ = json.Unmarshal(agents["build"], &build)
	permission := map[string]json.RawMessage{}
	_ = json.Unmarshal(build["permission"], &permission)
	task := map[string]string{}
	_ = json.Unmarshal(permission["task"], &task)
	if task["planner"] != "allow" {
		t.Fatalf("planner task permission = %q, want allow", task["planner"])
	}
}

// AC6 no-clobber: an existing project's legacy agent definitions survive.
func TestMergeOpenCodeAgents_NeverClobbersExistingLegacyAgents(t *testing.T) {
	existing, _ := json.Marshal(map[string]any{
		"big-thinker": map[string]string{"description": "user-authored", "prompt": "keep me"},
	})
	raw := map[string]json.RawMessage{"agent": existing}
	if !mergeOpenCodeAgents(raw) {
		t.Fatal("merge must add planner to a legacy config")
	}
	agents := plannerAgentsOf(t, raw)
	var bt map[string]string
	_ = json.Unmarshal(agents["big-thinker"], &bt)
	if bt["description"] != "user-authored" || bt["prompt"] != "keep me" {
		t.Fatalf("existing legacy agent was rewritten: %#v", bt)
	}
	if _, ok := agents["planner"]; !ok {
		t.Fatal("planner must still be added alongside the legacy entries")
	}
}
