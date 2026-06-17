package gates

import (
	"testing"
)

// TestRetainFindings_NoFilterRetainsAll verifies nil filter retains all.
func TestRetainFindings_NoFilterRetainsAll(t *testing.T) {
	f := []gitleaksFinding{
		{RuleID: "rule-a", File: "a.go"},
		{RuleID: "rule-b", File: "b.go"},
	}
	kept := retainFindings(f, nil, nil)
	if len(kept) != 2 {
		t.Fatalf("nil filter must retain all findings, got %v", kept)
	}
}

// TestRetainFindings_UnmatchedAllowlistIsIgnored verifies noop allowlist.
func TestRetainFindings_UnmatchedAllowlistIsIgnored(t *testing.T) {
	f := []gitleaksFinding{{RuleID: "aws-key", File: "secret.go"}}
	kept := retainFindings(f, []string{"some-other-rule"}, nil)
	if len(kept) != 1 {
		t.Fatalf("unmatched allowlist must not suppress finding, got %v", kept)
	}
}

// TestAllowlisted_ExactRuleID exercises rule-ID equality matching.
func TestAllowlisted_ExactRuleID(t *testing.T) {
	if !allowlisted("my-rule", "f.go", []string{"my-rule"}) {
		t.Fatal("exact rule ID must match")
	}
	if allowlisted("other-rule", "f.go", []string{"my-rule"}) {
		t.Fatal("different rule ID must not match")
	}
}

// TestAllowlisted_GlobPath exercises filepath.Match path-glob semantics.
func TestAllowlisted_GlobPath(t *testing.T) {
	if !allowlisted("rule", "testdata/secret.go", []string{"testdata/*"}) {
		t.Fatal("glob must match matching path")
	}
	if allowlisted("rule", "src/secret.go", []string{"testdata/*"}) {
		t.Fatal("glob must not match non-matching path")
	}
}

// TestAllowlisted_EmptyAllowlistAlwaysFalse confirms empty list never suppresses.
func TestAllowlisted_EmptyAllowlistAlwaysFalse(t *testing.T) {
	if allowlisted("any-rule", "any.go", nil) {
		t.Fatal("empty allowlist must never suppress")
	}
}

// TestRetainFindings_DeduplicatesSameFileSameRule verifies deduplication.
func TestRetainFindings_DeduplicatesSameFileSameRule(t *testing.T) {
	f := []gitleaksFinding{
		{RuleID: "aws-key", File: "a.go"},
		{RuleID: "aws-key", File: "a.go"},
	}
	kept := retainFindings(f, nil, nil)
	if len(kept) != 1 {
		t.Fatalf("duplicate findings must be deduped, got %v", kept)
	}
}

// TestRetainFindings_DetailLineFormat verifies the "file: rule X" line format.
func TestRetainFindings_DetailLineFormat(t *testing.T) {
	f := []gitleaksFinding{{RuleID: "my-rule", File: "src/config.go"}}
	kept := retainFindings(f, nil, nil)
	if len(kept) != 1 {
		t.Fatalf("expected 1 line, got %v", kept)
	}
	expected := "src/config.go: rule my-rule"
	if kept[0] != expected {
		t.Fatalf("expected %q, got %q", expected, kept[0])
	}
}
