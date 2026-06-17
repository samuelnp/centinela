package roadmap

import (
	"os"
	"testing"
)

// TestWriteArtifact_InvalidFeaturesArray returns error.
func TestWriteArtifact_InvalidFeaturesArray(t *testing.T) {
	d := t.TempDir()
	orig, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(orig) }) //nolint:errcheck
	os.Chdir(d)                          //nolint:errcheck
	os.MkdirAll(".workflow", 0755)       //nolint:errcheck
	p := RoadmapAnalysisFile
	// Write a file with a non-array features value so writeFeatureArray fails.
	// We'll call writeArtifact directly with features as a scalar
	top := map[string]interface{}{
		"features": "not-an-array",
	}
	// Build the RawMessage map manually since we need to trigger the error path
	// inside writeFeatureArray (json.Unmarshal of features)
	import_json := `{"features":"not-an-array"}`
	os.WriteFile(p, []byte(import_json), 0644) //nolint:errcheck
	// appendFeatureEntry reads the file and calls writeArtifact with raw "features":"not-an-array"
	// which causes writeFeatureArray's Unmarshal to fail
	entry, _ := compactBytes(AnalysisFeature{Name: "x"})
	if err := appendFeatureEntry(p, entry); err == nil {
		t.Error("expected error when features is not an array")
	}
	_ = top
}

// TestAppendLine_NoTrailingNewline adds newline separator when content lacks one.
func TestAppendLine_NoTrailingNewline(t *testing.T) {
	d := t.TempDir()
	orig, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(orig) }) //nolint:errcheck
	os.Chdir(d)                          //nolint:errcheck
	os.MkdirAll(".workflow", 0755)       //nolint:errcheck
	// File with no trailing newline
	p := RoadmapAnalysisMarkdown
	os.WriteFile(p, []byte("existing"), 0644) //nolint:errcheck
	appendLine(p, "- bullet")                 //nolint:errcheck
	data, _ := os.ReadFile(p)
	s := string(data)
	// Must have a newline between existing and bullet
	if s != "existing\n- bullet\n" {
		t.Errorf("unexpected append result: %q", s)
	}
}
