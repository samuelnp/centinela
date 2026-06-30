package setup

import (
	"encoding/json"
	"strings"
	"testing"
)

func rawWith(provider string) map[string]json.RawMessage {
	raw := map[string]json.RawMessage{}
	if provider != "" {
		raw["provider"] = json.RawMessage(provider)
	}
	return raw
}

func providerKeys(t *testing.T, raw map[string]json.RawMessage) map[string]json.RawMessage {
	t.Helper()
	var providers map[string]json.RawMessage
	if err := json.Unmarshal(raw["provider"], &providers); err != nil {
		t.Fatalf("unmarshal provider: %v", err)
	}
	return providers
}

// mergeProvider: nil is a no-op; the first add returns true and writes the managed
// key; an idempotent re-add returns false; a real endpoint change rewrites the
// owned block (the only changed=true trigger after the key exists).
func TestMergeProvider(t *testing.T) {
	if mergeProvider(map[string]json.RawMessage{}, nil) {
		t.Fatal("nil local must be a no-op")
	}
	lp := &LocalProvider{Provider: "ollama", Endpoint: "http://localhost:11434/v1", Model: "qwen2.5-coder"}

	raw := map[string]json.RawMessage{}
	if !mergeProvider(raw, lp) {
		t.Fatal("first add should change")
	}
	if _, ok := providerKeys(t, raw)["ollama"]; !ok {
		t.Fatal("managed ollama key missing")
	}
	if mergeProvider(raw, lp) {
		t.Fatal("idempotent re-add should not change")
	}

	changed := &LocalProvider{Provider: "ollama", Endpoint: "http://localhost:11500/v1", Model: "qwen2.5-coder"}
	if !mergeProvider(raw, changed) {
		t.Fatal("endpoint change should rewrite")
	}
	if got := string(providerKeys(t, raw)["ollama"]); !strings.Contains(got, "11500") {
		t.Fatalf("updated block missing new endpoint: %s", got)
	}
}
