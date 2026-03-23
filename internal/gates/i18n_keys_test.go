package gates

import "testing"

func TestFlatKeysAndCompare(t *testing.T) {
	keys := flatKeys(map[string]interface{}{"a": map[string]interface{}{"b": "x"}, "c": "y"}, "")
	if !keys["a.b"] || !keys["c"] {
		t.Fatalf("flatKeys missing keys: %#v", keys)
	}
	resPass := compareKeysets(map[string]map[string]bool{"en": {"k": true}, "es": {"k": true}}, []string{"en", "es"})
	if resPass.Status != Pass {
		t.Fatalf("expected pass, got %v", resPass.Status)
	}
	resFail := compareKeysets(map[string]map[string]bool{"en": {"k": true}, "es": {}}, []string{"en", "es"})
	if resFail.Status != Fail || len(resFail.Details) == 0 {
		t.Fatalf("expected fail with details, got %+v", resFail)
	}
}
