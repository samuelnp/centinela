package roadmap

import (
	"os"
	"strings"
	"testing"
)

// phaseOpsChdir seeds .workflow/roadmap.json plus empty analysis/quality
// artifacts in a fresh temp cwd, for --force prune tests that read those files.
func phaseOpsChdir(t *testing.T, body string) {
	t.Helper()
	crudChdir(t, body)
	os.WriteFile(RoadmapAnalysisFile, []byte(`{"role":"senior-product-manager","features":[]}`), 0o644)                 //nolint:errcheck
	os.WriteFile(RoadmapQualityFile, []byte(`{"role":"roadmap-quality-evaluator","threshold":9,"features":[]}`), 0o644) //nolint:errcheck
	os.WriteFile(RoadmapAnalysisMarkdown, []byte("# analysis\n"), 0o644)                                                //nolint:errcheck
	os.WriteFile(RoadmapQualityMarkdown, []byte("# quality\n"), 0o644)                                                  //nolint:errcheck
}

// phaseOrderNames returns the on-disk phase-name order at path.
func phaseOrderNames(t *testing.T, path string) []string {
	t.Helper()
	doc, err := readRawRoadmap(path)
	if err != nil {
		t.Fatalf("readRawRoadmap: %v", err)
	}
	names := make([]string, 0, len(doc.phases))
	for i := range doc.phases {
		n, err := phaseName(doc.phaseBytes(i))
		if err != nil {
			t.Fatalf("phaseName: %v", err)
		}
		names = append(names, n)
	}
	return names
}

// wantErr fails unless err is non-nil and its message contains sub.
func wantErr(t *testing.T, err error, sub string) {
	t.Helper()
	if err == nil || !strings.Contains(err.Error(), sub) {
		t.Fatalf("want error containing %q, got %v", sub, err)
	}
}
