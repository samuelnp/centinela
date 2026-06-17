package roadmap

import (
	"os"
	"strings"
	"testing"
)

// TestProvenanceBullet_WithSource formats source/feature/role correctly.
func TestProvenanceBullet_WithSource(t *testing.T) {
	f := &BacklogFinding{
		Name:       "my-slug",
		Source:     &Source{Feature: "dfrc", Role: "senior-engineer"},
		DeferredAt: "2026-01-01T00:00:00Z",
	}
	got := provenanceBullet("my-slug", f)
	if !strings.Contains(got, "dfrc/senior-engineer") {
		t.Errorf("source not in bullet: %s", got)
	}
	if !strings.Contains(got, "2026-01-01T00:00:00Z") {
		t.Errorf("deferredAt not in bullet: %s", got)
	}
	if !strings.Contains(got, "my-slug") {
		t.Errorf("slug not in bullet: %s", got)
	}
}

// TestProvenanceBullet_NoSource uses "unknown" when Source is nil (regression).
func TestProvenanceBullet_NoSource(t *testing.T) {
	f := &BacklogFinding{Name: "no-src", Source: nil, DeferredAt: "2026-01-02T00:00:00Z"}
	got := provenanceBullet("no-src", f)
	if !strings.Contains(got, "unknown") {
		t.Errorf("nil source must yield 'unknown' in bullet: %s", got)
	}
}

// TestProvenanceBullet_SourceFeatureOnly no role produces feature-only source.
func TestProvenanceBullet_SourceFeatureOnly(t *testing.T) {
	f := &BacklogFinding{Name: "x", Source: &Source{Feature: "my-feat"}, DeferredAt: "t"}
	got := provenanceBullet("x", f)
	if !strings.Contains(got, "my-feat") {
		t.Errorf("feature not in bullet: %s", got)
	}
}

// TestAppendFeatureEntry_AppendsToExistingArray adds an entry to features.
func TestAppendFeatureEntry_AppendsToExistingArray(t *testing.T) {
	d := t.TempDir()
	orig, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(orig) })                                                                                  //nolint:errcheck
	os.Chdir(d)                                                                                                           //nolint:errcheck
	os.MkdirAll(".workflow", 0755)                                                                                        //nolint:errcheck
	os.WriteFile(RoadmapAnalysisFile, []byte(`{"role":"senior-product-manager","features":[{"name":"existing"}]}`), 0644) //nolint:errcheck
	entry, _ := compactBytes(AnalysisFeature{Name: "new-slug"})
	if err := appendFeatureEntry(RoadmapAnalysisFile, entry); err != nil {
		t.Fatalf("appendFeatureEntry: %v", err)
	}
	data, _ := os.ReadFile(RoadmapAnalysisFile)
	if !strings.Contains(string(data), "existing") {
		t.Error("prior entry must be preserved")
	}
	if !strings.Contains(string(data), "new-slug") {
		t.Error("new entry must be appended")
	}
}

// TestAppendFeatureEntry_MissingFile returns error.
func TestAppendFeatureEntry_MissingFile(t *testing.T) {
	entry, _ := compactBytes(AnalysisFeature{Name: "x"})
	if err := appendFeatureEntry("/nonexistent/path.json", entry); err == nil {
		t.Error("expected error for missing artifact file")
	}
}

// TestAppendFeatureEntry_CorruptJSON returns error.
func TestAppendFeatureEntry_CorruptJSON(t *testing.T) {
	d := t.TempDir()
	orig, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(orig) })                    //nolint:errcheck
	os.Chdir(d)                                             //nolint:errcheck
	os.MkdirAll(".workflow", 0755)                          //nolint:errcheck
	os.WriteFile(RoadmapAnalysisFile, []byte("{bad"), 0644) //nolint:errcheck
	entry, _ := compactBytes(AnalysisFeature{Name: "x"})
	if err := appendFeatureEntry(RoadmapAnalysisFile, entry); err == nil {
		t.Error("expected error for corrupt JSON")
	}
}
