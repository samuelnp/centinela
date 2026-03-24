package migration

import (
	"strings"
	"testing"
)

func TestMergePreservesKeepAndCustom(t *testing.T) {
	current := "# Doc\n\n<!-- centinela:keep:start:notes -->\nhello\n<!-- centinela:keep:end:notes -->\n\n## Extra\nvalue\n"
	tpl := "# Doc\n\n<!-- centinela:keep:start:notes -->\ndefault\n<!-- centinela:keep:end:notes -->\n\n## Core\nok\n"
	out, keep, custom := mergeContent(current, tpl)
	if keep != 1 || custom != 1 {
		t.Fatalf("expected keep=1 custom=1, got %d %d", keep, custom)
	}
	if !strings.Contains(out, "hello") || !strings.Contains(out, "## Preserved Custom Sections") {
		t.Fatal("expected preserved content in merge output")
	}
}
