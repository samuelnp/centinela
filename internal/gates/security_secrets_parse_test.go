package gates

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/gitdiff"
)

// writeReport writes content to a temp file and returns its path.
func writeReport(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "report*.json")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = f.WriteString(content)
	_ = f.Close()
	return f.Name()
}

// TestReadSecretsReport_EmptyFileIsClean verifies empty report = no findings.
func TestReadSecretsReport_EmptyFileIsClean(t *testing.T) {
	p := writeReport(t, "")
	got, err := readSecretsReport(p)
	if err != nil || got != nil {
		t.Fatalf("empty file must yield nil findings, got %v / %v", got, err)
	}
}

// TestReadSecretsReport_HappyFindings decodes a well-formed findings array.
func TestReadSecretsReport_HappyFindings(t *testing.T) {
	json := `[{"RuleID":"aws-access-key","File":"config.go"}]`
	p := writeReport(t, json)
	got, err := readSecretsReport(p)
	if err != nil || len(got) != 1 || got[0].RuleID != "aws-access-key" {
		t.Fatalf("unexpected: err=%v findings=%v", err, got)
	}
}

// TestReadSecretsReport_MalformedJSONIsError verifies non-JSON -> error.
func TestReadSecretsReport_MalformedJSONIsError(t *testing.T) {
	p := writeReport(t, "not json at all")
	_, err := readSecretsReport(p)
	if err == nil {
		t.Fatal("expected parse error for malformed JSON")
	}
}

// TestReadSecretsReport_MissingFileIsClean verifies missing report = clean.
func TestReadSecretsReport_MissingFileIsClean(t *testing.T) {
	got, err := readSecretsReport(filepath.Join(t.TempDir(), "missing.json"))
	if err != nil || got != nil {
		t.Fatalf("missing file must yield nil findings, got %v / %v", got, err)
	}
}

// TestRetainFindings_AllowlistByRuleID verifies rule-ID suppression.
func TestRetainFindings_AllowlistByRuleID(t *testing.T) {
	f := []gitleaksFinding{{RuleID: "generic-api-key", File: "config.go"}}
	kept := retainFindings(f, []string{"generic-api-key"}, nil)
	if len(kept) != 0 {
		t.Fatalf("allowlisted finding should be suppressed, got %v", kept)
	}
}

// TestRetainFindings_AllowlistByPathGlob verifies path-glob suppression.
func TestRetainFindings_AllowlistByPathGlob(t *testing.T) {
	f := []gitleaksFinding{{RuleID: "generic-api-key", File: "testdata/secret.go"}}
	kept := retainFindings(f, []string{"testdata/*"}, nil)
	if len(kept) != 0 {
		t.Fatalf("path-glob allowlisted finding should be suppressed, got %v", kept)
	}
}

// TestRetainFindings_DiffFilterDropsOutOfDiff verifies diff-filter exclusion.
func TestRetainFindings_DiffFilterDropsOutOfDiff(t *testing.T) {
	f := []gitleaksFinding{{RuleID: "aws-key", File: "unchanged.go"}}
	filter := gitdiff.NewSet([]string{"changed.go"})
	kept := retainFindings(f, nil, filter)
	if len(kept) != 0 {
		t.Fatalf("out-of-diff finding must be filtered, got %v", kept)
	}
}

// TestRetainFindings_DiffFilterIncludesInDiff verifies in-diff files pass.
func TestRetainFindings_DiffFilterIncludesInDiff(t *testing.T) {
	f := []gitleaksFinding{{RuleID: "aws-key", File: "changed.go"}}
	filter := gitdiff.NewSet([]string{"changed.go"})
	kept := retainFindings(f, nil, filter)
	if len(kept) != 1 {
		t.Fatalf("in-diff finding must be retained, got %v", kept)
	}
}
