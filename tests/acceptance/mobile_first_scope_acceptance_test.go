package acceptance_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestMobileFirstScopedToUXUISpecialist asserts that the optional
// `mobileFirst` field appears only in `.workflow/<feature>-ux-ui-specialist.json`
// evidence files. evidence-contract.md Rule 6 says the field must be omitted
// for every other role; this guard catches drift from prompt briefs or future
// CLI defaults that silently leak the field into non-UX evidence.
func TestMobileFirstScopedToUXUISpecialist(t *testing.T) {
	root := filepath.Join("..", "..", ".workflow")
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatalf("read .workflow: %v", err)
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		if strings.Contains(e.Name(), "ux-ui-specialist") {
			continue
		}
		path := filepath.Join(root, e.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		var raw map[string]json.RawMessage
		if err := json.Unmarshal(data, &raw); err != nil {
			continue
		}
		if _, ok := raw["mobileFirst"]; ok {
			t.Errorf("%s: non-UX evidence carries mobileFirst (contract Rule 6: omit)", e.Name())
		}
	}
}
