package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

// configSwallow records one function that obtains a config from config.Load and
// then HIDES a failure, by either of the two idioms this codebase has actually
// shipped. Both produce a config that no successful load created, and every such
// config resolves strict — but only because config.ResolvedByLoad enforces it at
// the decision. This scan keeps the intent visible at the call site too.
type configSwallow struct {
	Func     string
	Discards bool // cfg, _ := config.Load()
	Rebinds  bool // cfg, err := config.Load(); if err != nil { cfg = &config.Config{} }
}

// scanConfigSwallows parses path and returns one entry per offending function.
// It matches config.Load EXACTLY — config.LoadForProfile is the sanctioned entry
// point and discarding ITS error is correct, because its fallback pins strict.
func scanConfigSwallows(t *testing.T, path string) []configSwallow {
	t.Helper()
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	var out []configSwallow
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}
		if s := scanFuncBody(fn); s.Discards || s.Rebinds {
			out = append(out, s)
		}
	}
	return out
}

// scanFuncBody classifies one function: does it call config.Load, and does it
// then discard the error or rebind the config to an empty struct?
func scanFuncBody(fn *ast.FuncDecl) configSwallow {
	s := configSwallow{Func: fn.Name.Name}
	var callsLoad, rebindsEmpty bool
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		assign, ok := n.(*ast.AssignStmt)
		if !ok {
			return true
		}
		for _, rhs := range assign.Rhs {
			if isConfigLoadCall(rhs) {
				callsLoad = true
				if hasBlankTrailing(assign.Lhs) {
					s.Discards = true
				}
			}
			if isEmptyConfigLiteral(rhs) {
				rebindsEmpty = true
			}
		}
		return true
	})
	s.Rebinds = callsLoad && rebindsEmpty
	s.Discards = s.Discards && callsLoad
	return s
}
