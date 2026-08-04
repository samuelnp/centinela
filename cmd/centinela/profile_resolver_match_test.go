package main

import (
	"go/ast"
	"go/token"
)

// isConfigLoadCall reports whether e is exactly a call to config.Load.
func isConfigLoadCall(e ast.Expr) bool {
	call, ok := e.(*ast.CallExpr)
	if !ok {
		return false
	}
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != "Load" {
		return false
	}
	pkg, ok := sel.X.(*ast.Ident)
	return ok && pkg.Name == "config"
}

// isEmptyConfigLiteral reports whether e is &config.Config{} or &Config{}.
func isEmptyConfigLiteral(e ast.Expr) bool {
	unary, ok := e.(*ast.UnaryExpr)
	if !ok || unary.Op != token.AND {
		return false
	}
	lit, ok := unary.X.(*ast.CompositeLit)
	if !ok || len(lit.Elts) != 0 {
		return false
	}
	if sel, ok := lit.Type.(*ast.SelectorExpr); ok {
		return sel.Sel.Name == "Config"
	}
	ident, ok := lit.Type.(*ast.Ident)
	return ok && ident.Name == "Config"
}

// hasBlankTrailing reports whether the last assigned name is the blank
// identifier — the error slot of `cfg, _ := config.Load()`.
func hasBlankTrailing(lhs []ast.Expr) bool {
	if len(lhs) == 0 {
		return false
	}
	ident, ok := lhs[len(lhs)-1].(*ast.Ident)
	return ok && ident.Name == "_"
}
