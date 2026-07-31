package docstring

import (
	"go/ast"
	"go/token"
)

// collector accumulates one file's findings into a shared Report.
type collector struct {
	path string
	fset *token.FileSet
	opts Options
	rep  *Report
}

// file walks a parsed file's top-level declarations.
func (c *collector) file(f *ast.File) {
	for _, d := range f.Decls {
		switch decl := d.(type) {
		case *ast.FuncDecl:
			c.funcDecl(decl)
		case *ast.GenDecl:
			c.genDecl(decl)
		}
	}
}

// funcDecl records an exported function or a method on an exported type.
// Methods on unexported types are unreachable from outside the package and are
// therefore not API surface.
func (c *collector) funcDecl(d *ast.FuncDecl) {
	if !ast.IsExported(d.Name.Name) {
		return
	}
	kind := "func"
	if d.Recv != nil {
		if !exportedReceiver(d.Recv) {
			return
		}
		kind = "method"
	}
	c.record(kind, d.Name.Name, d.Pos(), d.Doc, nil)
}

// genDecl records the exported specs of a const, var or type declaration,
// letting a doc comment on the block cover every spec inside it.
func (c *collector) genDecl(d *ast.GenDecl) {
	kind, ok := declKind(d.Tok)
	if !ok {
		return
	}
	for _, s := range d.Specs {
		switch spec := s.(type) {
		case *ast.ValueSpec:
			for _, n := range spec.Names {
				if ast.IsExported(n.Name) {
					c.record(kind, n.Name, n.Pos(), firstDoc(spec.Doc, d.Doc), spec.Comment)
				}
			}
		case *ast.TypeSpec:
			if ast.IsExported(spec.Name.Name) {
				c.record(kind, spec.Name.Name, spec.Pos(), firstDoc(spec.Doc, d.Doc), spec.Comment)
				c.members(spec)
			}
		}
	}
}

// declKind maps a GenDecl token to its reported kind.
func declKind(tok token.Token) (string, bool) {
	switch tok {
	case token.CONST:
		return "const", true
	case token.VAR:
		return "var", true
	case token.TYPE:
		return "type", true
	}
	return "", false
}
