package migration

import "testing"

func TestExtractKeepBlocksIgnoresUnclosed(t *testing.T) {
	in := "<!-- centinela:keep:start:x -->\nvalue"
	blocks := extractKeepBlocks(in)
	if len(blocks) != 0 {
		t.Fatal("expected unclosed keep block to be ignored")
	}
}

func TestReplaceKeepBlocksFallbackBody(t *testing.T) {
	tpl := "<!-- centinela:keep:start:x -->\ndef\n<!-- centinela:keep:end:x -->"
	out, kept := replaceKeepBlocks(tpl, map[string]string{})
	if kept != 0 || out == "" {
		t.Fatal("expected template block retained when no preserved content")
	}
}
