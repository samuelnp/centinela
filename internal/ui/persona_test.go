package ui

import (
	"strings"
	"testing"
)

func TestPersonaLabelIsFixedEmojiPrefix(t *testing.T) {
	if personaLabel() != " 🛡️👁️ " {
		t.Fatal("persona label should be fixed emoji pair")
	}
}

func TestRenderSystemLineIncludesPersona(t *testing.T) {
	line := renderSystemLine("CLI", "hello", toneError)
	if !strings.Contains(line, "🛡️👁️") {
		t.Fatal("render line should include persona label")
	}
	if !strings.Contains(line, "CLI") || !strings.Contains(line, "hello") {
		t.Fatal("render line should keep channel and message")
	}
}
