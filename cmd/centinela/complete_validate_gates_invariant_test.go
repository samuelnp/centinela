package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

// gatesSourceFile is the file whose header states the invariant this test
// mechanizes: "profiles scale process, never proof".
const gatesSourceFile = "complete_validate_gates.go"

// TestValidateGatesHasNoProfileIdentifier is a SOURCE-LEVEL guard, deliberately
// coarser than a behavior test: a future edit that reads, branches on, or even
// mentions a profile inside the validate-step gate runner fails here rather than
// depending on a reviewer noticing. Identifiers are taken from the parsed AST,
// so the prose in the file's own comments (which says "profile" repeatedly)
// cannot trip it — only real code can.
func TestValidateGatesHasNoProfileIdentifier(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), gatesSourceFile, nil, 0)
	if err != nil {
		t.Fatalf("parse %s: %v", gatesSourceFile, err)
	}
	ast.Inspect(file, func(n ast.Node) bool {
		id, ok := n.(*ast.Ident)
		if ok && strings.Contains(strings.ToLower(id.Name), "profile") {
			t.Errorf("%s must contain no profile identifier, found %q at %v — "+
				"verification is CONSTANT across every profile; profiles scale "+
				"process, never proof", gatesSourceFile, id.Name, id.Pos())
		}
		return true
	})
}

// TestValidateGatesInvariantGuardCatchesAProfileBranch proves the guard above
// can fail: the same AST scan over a synthetic source that DOES branch on a
// profile must report it. Without this, a broken scan would pass silently
// forever and the invariant would be unguarded while looking guarded.
func TestValidateGatesInvariantGuardCatchesAProfileBranch(t *testing.T) {
	const src = `package main
func runValidateGates(profile string) error { if profile == "guided" { return nil }; return nil }`
	file, err := parser.ParseFile(token.NewFileSet(), "synthetic.go", src, 0)
	if err != nil {
		t.Fatalf("parse synthetic: %v", err)
	}
	found := false
	ast.Inspect(file, func(n ast.Node) bool {
		if id, ok := n.(*ast.Ident); ok && strings.Contains(strings.ToLower(id.Name), "profile") {
			found = true
		}
		return true
	})
	if !found {
		t.Fatal("guard failed to detect a profile identifier in a source that has one")
	}
}
