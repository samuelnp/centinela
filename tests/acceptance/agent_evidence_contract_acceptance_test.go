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
	for role, path := range rolePrompts {
		b, err := os.ReadFile(repoRel(path))
		if err != nil {
			t.Fatalf("read %s prompt: %v", role, err)
		}
		s := string(b)
		if !strings.Contains(s, "evidence-contract.md") {
			t.Fatalf("%s prompt missing reference to evidence-contract.md", role)
		}
		if !strings.Contains(s, "JSON skeleton") {
			t.Fatalf("%s prompt missing JSON skeleton section", role)
		}
		if !strings.Contains(s, `"role": "`+role+`"`) {
			t.Fatalf("%s prompt JSON skeleton missing role field with value %q", role, role)
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
	b, err := os.ReadFile(repoRel(rolePrompts["ux-ui-specialist"]))
	if err != nil {
		t.Fatalf("read ux prompt: %v", err)
	}
	s := string(b)
	if !strings.Contains(s, `"mobileFirst": true`) {
		t.Fatal("ux prompt missing mobileFirst: true")
	}
	for _, tag := range []string{
		"mobile-first", "visual-hierarchy", "typography-hierarchy",
		"responsive-layout", "loading-state", "empty-state",
		"error-state", "motion-and-reduced-motion",
	} {
		if !strings.Contains(s, tag) {
			t.Fatalf("ux prompt missing required tag %q", tag)
		}
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
