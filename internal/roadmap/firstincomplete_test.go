package roadmap

import (
	"os"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

func chdirRoadmapTemp(t *testing.T) {
	t.Helper()
	d := t.TempDir()
	o, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(o) }) //nolint:errcheck
	if err := os.Chdir(d); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(workflow.WorkflowDir, 0o755); err != nil {
		t.Fatal(err)
	}
}

func seedDone(t *testing.T, name string) {
	t.Helper()
	wf := workflow.New(name)
	wf.CurrentStep = "done"
	if err := workflow.Save(wf); err != nil {
		t.Fatalf("save %s: %v", name, err)
	}
}

// FirstIncomplete crosses phases; nil/all-done yield ("", false).
func TestFirstIncomplete_Internal(t *testing.T) {
	chdirRoadmapTemp(t)
	seedDone(t, "p0a")
	r := &Roadmap{Phases: []Phase{
		{Name: "P0", Features: []Feature{{Name: "p0a"}}},
		{Name: "P1", Features: []Feature{{Name: "p1a"}}},
	}}
	if name, ok := FirstIncomplete(r); !ok || name != "p1a" {
		t.Fatalf("want (p1a,true), got (%q,%v)", name, ok)
	}
	if name, ok := FirstIncomplete(nil); ok || name != "" {
		t.Fatalf("nil want (\"\",false), got (%q,%v)", name, ok)
	}
	seedDone(t, "p1a")
	if name, ok := FirstIncomplete(r); ok || name != "" {
		t.Fatalf("all-done want (\"\",false), got (%q,%v)", name, ok)
	}
}

// FirstNotDone: planned feature → true; done feature → ("", false).
func TestFirstNotDone_Internal(t *testing.T) {
	chdirRoadmapTemp(t)
	if name, ok := FirstNotDone("fresh"); !ok || name != "fresh" {
		t.Fatalf("planned want (fresh,true), got (%q,%v)", name, ok)
	}
	seedDone(t, "fin")
	if name, ok := FirstNotDone("fin"); ok || name != "" {
		t.Fatalf("done want (\"\",false), got (%q,%v)", name, ok)
	}
}
