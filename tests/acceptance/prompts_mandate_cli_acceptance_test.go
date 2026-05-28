package acceptance_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Acceptance for specs/evidence-cli.feature scenario:
// "Agent prompts forbid hand-written JSON".

var promptRoles = []string{
	"big-thinker", "feature-specialist", "senior-engineer",
	"ux-ui-specialist", "qa-senior", "validation-specialist",
	"gatekeeper", "documentation-generator", "production-readiness",
}

// mirrorPromptName maps a source basename to its mirror basename when
// the two differ. Production-readiness mirrors `*.md.template`.
var mirrorPromptName = map[string]string{
	"production-readiness-prompt.md": "production-readiness-prompt.md.template",
}

// artifactPrompts must also reference `centinela artifact new` because
// their workflow saves a templated .workflow companion stub.
var artifactPrompts = map[string]struct{}{
	"gatekeeper-prompt.md":              {},
	"qa-senior-prompt.md":               {},
	"production-readiness-prompt.md":    {},
	"documentation-generator-prompt.md": {},
}

type promptCase struct {
	label string
	path  string
	base  string
}

func collectPromptCases(t *testing.T) []promptCase {
	t.Helper()
	cases := make([]promptCase, 0, len(promptRoles)*2)
	for _, role := range promptRoles {
		base := role + "-prompt.md"
		src := filepath.Join("..", "..", "docs", "architecture", base)
		mirrorBase := base
		if alt, ok := mirrorPromptName[base]; ok {
			mirrorBase = alt
		}
		mir := filepath.Join("..", "..", "internal", "scaffold", "assets",
			"docs", "architecture", mirrorBase)
		cases = append(cases,
			promptCase{"source/" + base, src, base},
			promptCase{"mirror/" + mirrorBase, mir, base},
		)
	}
	return cases
}

func TestPromptsMandateEvidenceCLI(t *testing.T) {
	for _, c := range collectPromptCases(t) {
		c := c
		t.Run(c.label, func(t *testing.T) {
			b, err := os.ReadFile(c.path)
			if err != nil {
				t.Fatalf("read %s: %v", c.path, err)
			}
			s := string(b)
			if !strings.Contains(s, "centinela evidence init") {
				t.Fatalf("%s missing `centinela evidence init` mandate", c.label)
			}
			if !strings.Contains(s, "centinela evidence set") &&
				!strings.Contains(s, "centinela evidence append") {
				t.Fatalf("%s missing `centinela evidence set|append` mandate", c.label)
			}
			assertNoForbiddenNearWorkflow(t, c.label, s)
			assertNoEmbeddedSkeleton(t, c.label, s)
			if _, ok := artifactPrompts[c.base]; ok {
				if !strings.Contains(s, "centinela artifact new") {
					t.Fatalf("%s missing `centinela artifact new` mandate", c.label)
				}
			}
		})
	}
}
