package acceptance_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Acceptance: specs/promote-orchestration-agents.feature
//
// Asserts the six promoted orchestration prompt files exist, contain the
// required section headings, are mirrored byte-identically into the
// scaffold tree, and respect the per-file line budget.

var promotedPromptFiles = []string{
	"big-thinker-prompt.md",
	"feature-specialist-prompt.md",
	"senior-engineer-prompt.md",
	"qa-senior-prompt.md",
	"ux-ui-specialist-prompt.md",
	"validation-specialist-prompt.md",
}

const promotedPromptLineBudget = 70

func docsArchPromptPath(name string) string {
	return filepath.Join("..", "..", "docs", "architecture", name)
}

func scaffoldPromptPath(name string) string {
	return filepath.Join("..", "..", "internal", "scaffold", "assets", "docs", "architecture", name)
}

func TestPromoteOrchestrationAgents_FilesExist(t *testing.T) {
	for _, name := range promotedPromptFiles {
		if _, err := os.Stat(docsArchPromptPath(name)); err != nil {
			t.Fatalf("missing promoted prompt %q: %v", name, err)
		}
	}
}

func TestPromoteOrchestrationAgents_RequiredSections(t *testing.T) {
	required := []string{"## Purpose", "## Prompt Template", "## Required Artifact"}
	for _, name := range promotedPromptFiles {
		data, err := os.ReadFile(docsArchPromptPath(name))
		if err != nil {
			t.Fatalf("read %q: %v", name, err)
		}
		s := string(data)
		for _, heading := range required {
			if !strings.Contains(s, heading) {
				t.Fatalf("%s missing required heading %q", name, heading)
			}
		}
	}
}

func TestPromoteOrchestrationAgents_MirrorByteIdentical(t *testing.T) {
	for _, name := range promotedPromptFiles {
		canonical, err := os.ReadFile(docsArchPromptPath(name))
		if err != nil {
			t.Fatalf("read canonical %q: %v", name, err)
		}
		mirror, err := os.ReadFile(scaffoldPromptPath(name))
		if err != nil {
			t.Fatalf("read scaffold mirror %q: %v", name, err)
		}
		if !bytes.Equal(canonical, mirror) {
			t.Fatalf("scaffold mirror drift for %q", name)
		}
	}
}

func TestPromoteOrchestrationAgents_LineBudget(t *testing.T) {
	for _, name := range promotedPromptFiles {
		data, err := os.ReadFile(docsArchPromptPath(name))
		if err != nil {
			t.Fatalf("read %q: %v", name, err)
		}
		lines := bytes.Count(data, []byte("\n"))
		if !bytes.HasSuffix(data, []byte("\n")) {
			lines++
		}
		if lines > promotedPromptLineBudget {
			t.Fatalf("%s has %d lines, exceeds budget of %d", name, lines, promotedPromptLineBudget)
		}
	}
}
