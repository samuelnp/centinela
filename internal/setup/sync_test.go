package setup

import (
	"os"
	"strings"
	"testing"
)

func TestBuildSyncPlanCreateMissing(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	plan, err := BuildSyncPlan("both")
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Items) != 4 {
		t.Fatalf("expected 4 setup items, got %d", len(plan.Items))
	}
}

func TestBuildSyncPlanManualReviewForCustomManagedFiles(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	os.MkdirAll(".opencode/plugins", 0755)                                 //nolint:errcheck
	os.WriteFile(".opencode/plugins/centinela.js", []byte("custom"), 0644) //nolint:errcheck
	os.WriteFile("AGENTS.md", []byte("custom agents"), 0644)               //nolint:errcheck
	plan, err := BuildSyncPlan("opencode")
	if err != nil {
		t.Fatal(err)
	}
	gotManual := 0
	for _, it := range plan.Items {
		if it.Action == SyncManualReview {
			gotManual++
		}
	}
	if gotManual != 2 {
		t.Fatalf("expected 2 manual-review items, got %d", gotManual)
	}
}

func TestApplySyncWritesManagedHeaders(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	plan, _ := BuildSyncPlan("opencode")
	if err := ApplySync(plan); err != nil {
		t.Fatal(err)
	}
	plug, _ := os.ReadFile(".opencode/plugins/centinela.js") //nolint:errcheck
	if !strings.HasPrefix(string(plug), "// centinela:managed-version=") {
		t.Fatal("expected plugin managed header")
	}
	agents, _ := os.ReadFile("AGENTS.md") //nolint:errcheck
	if !strings.HasPrefix(string(agents), "<!-- centinela:managed-version=") {
		t.Fatal("expected agents managed header")
	}
}
