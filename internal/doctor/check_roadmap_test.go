package doctor

import (
	"os"
	"strings"
	"testing"
)

func TestRoadmapCheckNoFileNotApplicable(t *testing.T) {
	repoFixture(t)
	d := roadmapCheck{}.Run(Context{})
	if d.Status != OK || !strings.Contains(d.Message, "not applicable") {
		t.Fatalf("missing roadmap.json must be OK/not-applicable, got %v %q", d.Status, d.Message)
	}
}

func TestRoadmapCheckInSyncOK(t *testing.T) {
	repoFixture(t)
	seedRoadmap(t, "Phase 1: Core")
	d := roadmapCheck{}.Run(Context{})
	if d.Status != OK {
		t.Fatalf("in-sync clean roadmap must be OK, got %v %q", d.Status, d.Message)
	}
}

func TestRoadmapCheckDriftError(t *testing.T) {
	repoFixture(t)
	seedRoadmap(t, "Phase 1: Core")
	writeFile(t, "ROADMAP.md", "# Roadmap\n\nhand-edited\n")
	d := roadmapCheck{}.Run(Context{})
	if d.Status != Error || d.Repair == nil || !d.Repair.Safe {
		t.Fatalf("drift must Error with safe repair, got %v repair=%v", d.Status, d.Repair)
	}
	if !strings.Contains(strings.Join(d.Details, " "), "out of sync") {
		t.Fatalf("details must mention drift: %v", d.Details)
	}
}

func TestRoadmapCheckGlyphError(t *testing.T) {
	repoFixture(t)
	seedRoadmap(t, "✅ Phase 0: Bootstrap")
	d := roadmapCheck{}.Run(Context{})
	if d.Status != Error {
		t.Fatalf("glyph phase must Error, got %v", d.Status)
	}
	joined := strings.Join(d.Details, " ")
	if !strings.Contains(joined, "Phase 0") || !strings.Contains(joined, "prefix") {
		t.Fatalf("details must name offending phase + prefix breakage: %v", d.Details)
	}
}

func TestRoadmapRepairStripsGlyphRegeneratesIdempotent(t *testing.T) {
	repoFixture(t)
	seedRoadmap(t, "✅ Phase 0: Bootstrap")
	writeFile(t, "ROADMAP.md", "stale\n")
	if err := repairRoadmap(); err != nil {
		t.Fatalf("repair: %v", err)
	}
	d := roadmapCheck{}.Run(Context{})
	if d.Status != OK {
		t.Fatalf("post-repair must be OK, got %v %q", d.Status, d.Message)
	}
	jsonBefore, _ := os.ReadFile(".workflow/roadmap.json")
	if strings.Contains(string(jsonBefore), "✅") {
		t.Fatal("glyph must be stripped from roadmap.json")
	}
	if err := repairRoadmap(); err != nil {
		t.Fatalf("idempotent repair: %v", err)
	}
	jsonAfter, _ := os.ReadFile(".workflow/roadmap.json")
	if string(jsonBefore) != string(jsonAfter) {
		t.Fatal("second repair must leave roadmap.json byte-identical")
	}
}
