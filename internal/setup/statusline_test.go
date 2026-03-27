package setup

import (
	"encoding/json"
	"testing"
)

func TestEnsureStatusLineIdempotent(t *testing.T) {
	settings := map[string]json.RawMessage{}
	if !ensureStatusLine(settings) {
		t.Fatal("expected first call to add statusLine")
	}
	if ensureStatusLine(settings) {
		t.Fatal("expected second call to be no-op")
	}
}
