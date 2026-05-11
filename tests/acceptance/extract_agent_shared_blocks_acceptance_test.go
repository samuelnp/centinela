package acceptance_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Acceptance: specs/extract-agent-shared-blocks.feature

const (
	agentInvocationFile = "agent-invocation.md"
	stackChecksFile     = "stack-checks-reference.md"
	prodReadinessTpl    = "production-readiness-prompt.md.template"
	gatekeeperFile      = "gatekeeper-prompt.md"
)

var promptsReferencingInvocation = []string{
	"gatekeeper-prompt.md",
	"edge-case-tester-prompt.md",
	"production-readiness-prompt.md.template",
	"big-thinker-prompt.md",
	"feature-specialist-prompt.md",
	"senior-engineer-prompt.md",
	"qa-senior-prompt.md",
	"ux-ui-specialist-prompt.md",
	"validation-specialist-prompt.md",
}

func archPath(name string) string {
	return filepath.Join("..", "..", "docs", "architecture", name)
}

func archMirrorPath(name string) string {
	return filepath.Join("..", "..", "internal", "scaffold", "assets", "docs", "architecture", name)
}

func TestExtractAgentSharedBlocks_SharedFilesExist(t *testing.T) {
	for _, name := range []string{agentInvocationFile, stackChecksFile} {
		if _, err := os.Stat(archPath(name)); err != nil {
			t.Fatalf("missing shared reference %q: %v", name, err)
		}
	}
}

func TestExtractAgentSharedBlocks_AgentInvocationDescribesContract(t *testing.T) {
	data, err := os.ReadFile(archPath(agentInvocationFile))
	if err != nil {
		t.Fatalf("read %q: %v", agentInvocationFile, err)
	}
	s := string(data)
	for _, want := range []string{"Agent", ".workflow/<feature>-<role>", "FEATURE_NAME"} {
		if !strings.Contains(s, want) {
			t.Fatalf("%s missing required mention of %q", agentInvocationFile, want)
		}
	}
}

func TestExtractAgentSharedBlocks_PromptsReferenceShared(t *testing.T) {
	for _, name := range promptsReferencingInvocation {
		data, err := os.ReadFile(archPath(name))
		if err != nil {
			t.Fatalf("read %q: %v", name, err)
		}
		if !strings.Contains(string(data), "agent-invocation.md") {
			t.Fatalf("%s does not reference agent-invocation.md", name)
		}
	}
}

func TestExtractAgentSharedBlocks_GatekeeperDecisionRulesRemoved(t *testing.T) {
	data, err := os.ReadFile(archPath(gatekeeperFile))
	if err != nil {
		t.Fatalf("read %q: %v", gatekeeperFile, err)
	}
	s := string(data)
	if strings.Contains(s, "## Decision Rules") {
		t.Fatalf("%s still contains the duplicate Decision Rules section", gatekeeperFile)
	}
	// SAFE / WARNING / BLOCKING decisions must still be expressed somewhere.
	for _, want := range []string{"SAFE", "WARNING", "BLOCKING"} {
		if !strings.Contains(s, want) {
			t.Fatalf("%s missing status word %q after Decision Rules removal", gatekeeperFile, want)
		}
	}
}

func TestExtractAgentSharedBlocks_StackMatrixMovedOut(t *testing.T) {
	data, err := os.ReadFile(archPath(prodReadinessTpl))
	if err != nil {
		t.Fatalf("read %q: %v", prodReadinessTpl, err)
	}
	s := string(data)
	if !strings.Contains(s, "stack-checks-reference.md") {
		t.Fatalf("%s does not reference stack-checks-reference.md", prodReadinessTpl)
	}
	// The four-language inline matrix should no longer be present together.
	languageHits := 0
	for _, lang := range []string{"**Go**:", "**TypeScript**:", "**Python**:", "**Ruby**:"} {
		if strings.Contains(s, lang) {
			languageHits++
		}
	}
	if languageHits >= 3 {
		t.Fatalf("%s still appears to contain the multi-language matrix (%d language entries found)", prodReadinessTpl, languageHits)
	}
}

func TestExtractAgentSharedBlocks_ScaffoldMirrorParity(t *testing.T) {
	affected := append([]string{agentInvocationFile, stackChecksFile}, promptsReferencingInvocation...)
	for _, name := range affected {
		canonical, err := os.ReadFile(archPath(name))
		if err != nil {
			t.Fatalf("read canonical %q: %v", name, err)
		}
		mirror, err := os.ReadFile(archMirrorPath(name))
		if err != nil {
			t.Fatalf("read scaffold mirror %q: %v", name, err)
		}
		if !bytes.Equal(canonical, mirror) {
			t.Fatalf("scaffold mirror drift for %q", name)
		}
	}
}
