package setup

import (
	"strings"
	"testing"
)

// mergeProvider no-clobber: a foreign block under the managed key (no managed
// marker) is preserved unchanged and the apply reports no change.
func TestMergeProviderNoClobberForeign(t *testing.T) {
	raw := rawWith(`{"ollama":{"npm":"hand-written"}}`)
	lp := &LocalProvider{Provider: "ollama", Endpoint: "http://x/v1", Model: "m"}
	if mergeProvider(raw, lp) {
		t.Fatal("foreign same-key block must not be clobbered")
	}
	if got := string(providerKeys(t, raw)["ollama"]); !strings.Contains(got, "hand-written") {
		t.Fatalf("foreign block changed: %s", got)
	}
}

// mergeProvider: a non-object foreign value under the managed key is not a valid
// managed block, so it is treated as foreign and left unclobbered (no change).
func TestMergeProviderForeignNonObject(t *testing.T) {
	raw := rawWith(`{"ollama":123}`)
	lp := &LocalProvider{Provider: "ollama", Endpoint: "http://x/v1", Model: "m"}
	if mergeProvider(raw, lp) {
		t.Fatal("non-object foreign value must not be clobbered")
	}
	if got := string(providerKeys(t, raw)["ollama"]); got != "123" {
		t.Fatalf("foreign value changed: %s", got)
	}
}

// mergeProvider add-alongside: a foreign provider under a DIFFERENT key is kept
// and the managed key is added beside it (owns only its own key).
func TestMergeProviderAddAlongside(t *testing.T) {
	raw := rawWith(`{"my-custom-provider":{"npm":"x"}}`)
	lp := &LocalProvider{Provider: "ollama", Endpoint: "http://x/v1", Model: "m"}
	if !mergeProvider(raw, lp) {
		t.Fatal("add alongside should change")
	}
	keys := providerKeys(t, raw)
	if _, ok := keys["my-custom-provider"]; !ok {
		t.Fatal("foreign provider dropped")
	}
	if _, ok := keys["ollama"]; !ok {
		t.Fatal("managed provider not added")
	}
}
