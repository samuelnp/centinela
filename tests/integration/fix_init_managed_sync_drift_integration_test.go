package integration_test

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/setup"
)

// Integration: applying the OpenCode sync plan writes the managed-version header
// and is idempotent — re-planning reports nothing to create or update.
func TestOpencodeSyncApplyIsIdempotent(t *testing.T) {
	t.Chdir(t.TempDir())
	first, err := setup.BuildSyncPlan("opencode")
	if err != nil {
		t.Fatal(err)
	}
	if err := setup.ApplySync(first); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile("AGENTS.md")
	if !strings.HasPrefix(string(b), "<!-- centinela:managed-version=") {
		t.Fatalf("AGENTS.md missing managed-version header, got:\n%s", string(b))
	}
	second, err := setup.BuildSyncPlan("opencode")
	if err != nil {
		t.Fatal(err)
	}
	for _, it := range second.Items {
		if it.Action == setup.SyncCreate || it.Action == setup.SyncUpdate {
			t.Errorf("not idempotent: %s still wants %q", it.Path, it.Action)
		}
	}
}
