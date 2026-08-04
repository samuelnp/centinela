package roadmap

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateQuality_EarlyErrorBranches(t *testing.T) {
	dir := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(dir)     //nolint:errcheck
	_ = os.MkdirAll(".workflow", 0755)

	r := &Roadmap{}

	// 1. Markdown missing.
	if err := ValidateQuality(r); err == nil || !strings.Contains(err.Error(), "markdown missing") {
		t.Fatalf("expected markdown-missing error, got %v", err)
	}

	// 2. Markdown present, JSON missing.
	_ = os.WriteFile(filepath.Join(".workflow", "roadmap-quality.md"), []byte("ok"), 0644)
	if err := ValidateQuality(r); err == nil || !strings.Contains(err.Error(), "json missing") {
		t.Fatalf("expected json-missing error, got %v", err)
	}

	// 3. Invalid JSON.
	_ = os.WriteFile(RoadmapQualityFile, []byte("{bad"), 0644)
	if err := ValidateQuality(r); err == nil || !strings.Contains(err.Error(), "invalid roadmap quality json") {
		t.Fatalf("expected invalid-json error, got %v", err)
	}

	// 4. Wrong role.
	_ = os.WriteFile(RoadmapQualityFile, []byte(`{"role":"nope"}`), 0644)
	if err := ValidateQuality(r); err == nil || !strings.Contains(err.Error(), "role must be") {
		t.Fatalf("expected wrong-role error, got %v", err)
	}

	// 5. A declared threshold other than 9 is no longer enforced — the field
	// still decodes so pre-deletion artifacts keep parsing, and the failure that
	// follows is about the ABSENT features array, never about the threshold.
	_ = os.WriteFile(RoadmapQualityFile,
		[]byte(`{"role":"roadmap-quality-evaluator","threshold":1}`), 0644)
	err := ValidateQuality(r)
	if err == nil || !strings.Contains(err.Error(), `"features" is missing`) {
		t.Fatalf("expected the missing-features shape error, got %v", err)
	}
	if strings.Contains(err.Error(), "threshold") {
		t.Fatalf("the threshold gate is deleted; it must not appear in %v", err)
	}
}
