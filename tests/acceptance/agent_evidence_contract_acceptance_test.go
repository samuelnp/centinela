package acceptance_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Acceptance: specs/add-agent-evidence-contract.feature

const contractPath = "docs/architecture/evidence-contract.md"

var rolePrompts = map[string]string{
	"big-thinker":              "docs/architecture/big-thinker-prompt.md",
	"feature-specialist":       "docs/architecture/feature-specialist-prompt.md",
	"senior-engineer":          "docs/architecture/senior-engineer-prompt.md",
	"qa-senior":                "docs/architecture/qa-senior-prompt.md",
	"ux-ui-specialist":         "docs/architecture/ux-ui-specialist-prompt.md",
	"validation-specialist":    "docs/architecture/validation-specialist-prompt.md",
	"documentation-specialist": "docs/architecture/documentation-generator-prompt.md",
}

func repoRel(p string) string { return filepath.Join("..", "..", p) }

func readContract(t *testing.T) string {
	t.Helper()
	b, err := os.ReadFile(repoRel(contractPath))
	if err != nil {
		t.Fatalf("read evidence-contract: %v", err)
	}
	return string(b)
}

func TestEvidenceContract_DocumentsSchemaAndAllRoles(t *testing.T) {
	body := readContract(t)
	for _, want := range []string{
		`"feature"`, `"step"`, `"role"`, `"status"`, `"generatedAt"`,
		`"inputs"`, `"outputs"`, `"edgeCases"`, `"mobileFirst"`, `"handoffTo"`,
		"big-thinker", "feature-specialist", "senior-engineer", "qa-senior",
		"ux-ui-specialist", "validation-specialist", "documentation-specialist",
		"RFC 3339",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("evidence-contract.md missing %q", want)
		}
	}
}

func TestEvidenceContract_PerRoleRulesPresent(t *testing.T) {
	body := readContract(t)
	rules := []string{
		"`docs/features/*.md`",
		"real implementation file",
		"at least one path under `tests/`",
		"mobileFirst",
		"motion-and-reduced-motion",
		"Exempt",
	}
	for _, want := range rules {
		if !strings.Contains(body, want) {
			t.Fatalf("evidence-contract.md missing rule statement %q", want)
		}
	}
}

func TestPromptsLinkToEvidenceContract(t *testing.T) {
	// Slice 3 (evidence-cli): the embedded JSON skeleton was removed from
	// every prompt and replaced with `centinela evidence schema <role>`
	// as the single source of truth. The prompt must still reference
	// evidence-contract.md and name the role in the schema invocation.
	for role, path := range rolePrompts {
		b, err := os.ReadFile(repoRel(path))
		if err != nil {
			t.Fatalf("read %s prompt: %v", role, err)
		}
		s := string(b)
		if !strings.Contains(s, "evidence-contract.md") {
			t.Fatalf("%s prompt missing reference to evidence-contract.md", role)
		}
		if !strings.Contains(s, "centinela evidence schema "+role) {
			t.Fatalf("%s prompt missing `centinela evidence schema %s` reference", role, role)
		}
	}
}

func TestPlanStepPromptsRequireSnapshotInputs(t *testing.T) {
	for _, role := range []string{"big-thinker", "feature-specialist"} {
		b, err := os.ReadFile(repoRel(rolePrompts[role]))
		if err != nil {
			t.Fatalf("read %s prompt: %v", role, err)
		}
		s := string(b)
		if !strings.Contains(s, "`docs/features/*.md`") || !strings.Contains(s, "snapshot") {
			t.Fatalf("%s prompt missing feature-doc snapshot rule", role)
		}
	}
}

func TestQASeniorPromptRequiresTestsAndEdgeCases(t *testing.T) {
	b, err := os.ReadFile(repoRel(rolePrompts["qa-senior"]))
	if err != nil {
		t.Fatalf("read qa-senior prompt: %v", err)
	}
	s := string(b)
	for _, want := range []string{"under `tests/`", "edge-cases.md"} {
		if !strings.Contains(s, want) {
			t.Fatalf("qa-senior prompt missing %q", want)
		}
	}
}

func TestUXPromptListsAllEightTagsAndMobileFirst(t *testing.T) {
	// Slice 3 (evidence-cli): the inline JSON skeleton (and the eight
	// enumerated tags it carried) was removed; the canonical tag list
	// now lives in docs/architecture/evidence-contract.md and is
	// enforced by `centinela evidence validate`. The prompt itself must
	// still surface mobileFirst, the validator's eight-tag rule, and
	// point readers at the contract.
	b, err := os.ReadFile(repoRel(rolePrompts["ux-ui-specialist"]))
	if err != nil {
		t.Fatalf("read ux prompt: %v", err)
	}
	s := string(b)
	if !strings.Contains(s, "mobileFirst") || !strings.Contains(s, "true") {
		t.Fatal("ux prompt missing mobileFirst:true mandate")
	}
	if !strings.Contains(s, "evidence-contract.md") {
		t.Fatal("ux prompt missing evidence-contract.md reference")
	}
	if !strings.Contains(s, "eight required UX tags") &&
		!strings.Contains(s, "eight UX tags") {
		t.Fatal("ux prompt missing eight-tag validator rule statement")
	}
}

func TestScaffoldMirrorParityForUpdatedPrompts(t *testing.T) {
	files := []string{
		"evidence-contract.md",
		"big-thinker-prompt.md", "feature-specialist-prompt.md",
		"senior-engineer-prompt.md", "qa-senior-prompt.md",
		"ux-ui-specialist-prompt.md", "validation-specialist-prompt.md",
		"documentation-generator-prompt.md",
	}
	for _, name := range files {
		live, err := os.ReadFile(filepath.Join("..", "..", "docs", "architecture", name))
		if err != nil {
			t.Fatalf("read live %s: %v", name, err)
		}
		mirror, err := os.ReadFile(filepath.Join("..", "..", "internal", "scaffold", "assets", "docs", "architecture", name))
		if err != nil {
			t.Fatalf("read scaffold %s: %v", name, err)
		}
		if string(live) != string(mirror) {
			t.Fatalf("scaffold mirror drift for %s", name)
		}
	}
}
