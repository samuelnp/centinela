package ui

import (
	"strings"
	"testing"
)

func TestPersonaFaceByTone(t *testing.T) {
	if personaFace(toneInfo) != "o_o" {
		t.Fatal("info persona should be o_o")
	}
	if personaFace(toneSuccess) != "^_^" {
		t.Fatal("success persona should be ^_^")
	}
	if personaFace(toneWarn) != "-_-" {
		t.Fatal("warn persona should be -_-")
	}
	if personaFace(toneError) != "ò_ó" {
		t.Fatal("error persona should be ò_ó")
	}
}

func TestRenderSystemLineIncludesPersona(t *testing.T) {
	line := renderSystemLine("CLI", "hello", toneError)
	if !strings.Contains(line, "CENTINELA says") {
		t.Fatal("render line should include persona label")
	}
	if !strings.Contains(line, "ò_ó") {
		t.Fatal("render line should include error face")
	}
	if !strings.Contains(line, "CLI") || !strings.Contains(line, "hello") {
		t.Fatal("render line should keep channel and message")
	}
}
