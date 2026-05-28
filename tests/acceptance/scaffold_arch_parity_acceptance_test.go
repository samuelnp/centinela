package acceptance_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// mirrorNameOverrides maps a source basename to a non-matching mirror
// basename. Seeded by evidence-cli Slice 3: the scaffold mirror for the
// production-readiness prompt is the templated variant `*.md.template`.
var mirrorNameOverrides = map[string]string{
	"production-readiness-prompt.md": "production-readiness-prompt.md.template",
}

// mirrorParityAllowlist tracks source files whose mirror diverges in
// content (or is absent) for reasons predating evidence-cli Slice 3.
// TODO(evidence-cli, Slice 3): track each entry under its own follow-up
// feature; do not extend this allowlist without a docs/features/*.md
// brief.
var mirrorParityAllowlist = map[string]string{
	"workflow-enforcement.md":        "TODO: not mirrored; track separately",
	"gatekeepers.md":                 "TODO: pre-existing scaffold drift, not caused by evidence-cli",
	"new-project-guide.md":           "TODO: pre-existing scaffold drift, not caused by evidence-cli",
	"testing-strategy.md":            "TODO: pre-existing scaffold drift, not caused by evidence-cli",
	"production-readiness-prompt.md": "TODO: source is project-instantiated; mirror is the generic *.md.template — content parity asserted by extract_agent_shared_blocks_acceptance_test.go on the `.template` pair",
}

func archDir() string { return filepath.Join("..", "..", "docs", "architecture") }

func archMirrorDir() string {
	return filepath.Join("..", "..", "internal", "scaffold", "assets",
		"docs", "architecture")
}

func resolveMirrorPath(srcBase string) string {
	if alt, ok := mirrorNameOverrides[srcBase]; ok {
		return filepath.Join(archMirrorDir(), alt)
	}
	return filepath.Join(archMirrorDir(), srcBase)
}

func firstDifferingLine(a, b []byte) int {
	la := bytes.Split(a, []byte("\n"))
	lb := bytes.Split(b, []byte("\n"))
	n := len(la)
	if len(lb) < n {
		n = len(lb)
	}
	for i := 0; i < n; i++ {
		if !bytes.Equal(la[i], lb[i]) {
			return i + 1
		}
	}
	if len(la) != len(lb) {
		return n + 1
	}
	return 0
}

func TestScaffoldArchitectureMirrorParity(t *testing.T) {
	entries, err := os.ReadDir(archDir())
	if err != nil {
		t.Fatalf("read arch dir: %v", err)
	}
	for _, ent := range entries {
		if ent.IsDir() || !strings.HasSuffix(ent.Name(), ".md") {
			continue
		}
		name := ent.Name()
		if reason, ok := mirrorParityAllowlist[name]; ok {
			t.Logf("skip %s (allowlist): %s", name, reason)
			continue
		}
		srcBytes, err := os.ReadFile(filepath.Join(archDir(), name))
		if err != nil {
			t.Fatalf("read source %s: %v", name, err)
		}
		mirrorPath := resolveMirrorPath(name)
		mirrorBytes, err := os.ReadFile(mirrorPath)
		if err != nil {
			t.Fatalf("missing mirror for %s (expected at %s): %v",
				name, mirrorPath, err)
		}
		if !bytes.Equal(srcBytes, mirrorBytes) {
			t.Fatalf("scaffold mirror drift for %s; first differing line: %d",
				name, firstDifferingLine(srcBytes, mirrorBytes))
		}
	}
}
