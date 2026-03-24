package migration

import (
	"strings"
	"testing"
)

func TestMergeAppendsMissingKeepBlocksSection(t *testing.T) {
	current := "<!-- centinela:keep:start:ghost -->\nabc\n<!-- centinela:keep:end:ghost -->"
	tpl := "# Doc\n"
	out, keep, custom := mergeContent(current, tpl)
	if keep != 1 || custom != 0 {
		t.Fatalf("expected keep=1 custom=0, got %d %d", keep, custom)
	}
	if !strings.Contains(out, "Preserved Keep Blocks") || !strings.Contains(out, "ghost") {
		t.Fatal("expected missing keep block appended")
	}
}
