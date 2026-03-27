package unit_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDocsCommandAndDocgenExist(t *testing.T) {
	cmdPath := filepath.Join("..", "..", "cmd", "centinela", "docs_generate.go")
	cmdData, err := os.ReadFile(cmdPath)
	if err != nil {
		t.Fatalf("read docs command: %v", err)
	}
	if !strings.Contains(string(cmdData), "Documentation generated") {
		t.Fatal("expected docs generate success output")
	}
	genPath := filepath.Join("..", "..", "internal", "docgen", "generate.go")
	genData, err := os.ReadFile(genPath)
	if err != nil {
		t.Fatalf("read docgen generate: %v", err)
	}
	if !strings.Contains(string(genData), "RenderHTML") {
		t.Fatal("expected html render usage")
	}
}
