package docstring

import (
	"go/ast"
	"go/token"
)

// members records the documented-ness of an exported type's members:
// interface methods always (they are the contract), struct fields only when
// Options.CheckFields is set.
func (c *collector) members(s *ast.TypeSpec) {
	switch t := s.Type.(type) {
	case *ast.InterfaceType:
		c.fields(t.Methods, "interface method")
	case *ast.StructType:
		if c.opts.CheckFields {
			c.fields(t.Fields, "field")
		}
	}
}

// fields records every exported named entry of a field list. Embedded entries
// carry no name and are skipped — they are documented at their own definition.
func (c *collector) fields(list *ast.FieldList, kind string) {
	if list == nil {
		return
	}
	for _, f := range list.List {
		for _, n := range f.Names {
			if ast.IsExported(n.Name) {
				c.record(kind, n.Name, n.Pos(), f.Doc, f.Comment)
			}
		}
	}
}

// record classifies one exported identifier: documented, exempt, or a
// violation. Documented wins over exempt so the exemption list only ever shows
// genuine opt-outs.
func (c *collector) record(kind, name string, pos token.Pos, doc, line *ast.CommentGroup) {
	c.rep.Inspected++
	at := c.fset.Position(pos).Line
	if documented(doc, name, c.opts.RequireNamePrefix) {
		return
	}
	if hasNodoc(doc, line) {
		c.rep.Exemptions = append(c.rep.Exemptions,
			Exemption{Path: c.path, Line: at, Kind: kind, Name: name})
		return
	}
	c.rep.Violations = append(c.rep.Violations,
		Violation{Path: c.path, Line: at, Kind: kind, Name: name})
}

// exportedReceiver reports whether a method's receiver base type is exported.
func exportedReceiver(recv *ast.FieldList) bool {
	if recv == nil || len(recv.List) == 0 {
		return false
	}
	return ast.IsExported(baseTypeName(recv.List[0].Type))
}

// baseTypeName unwraps pointer and generic receivers down to the type name.
func baseTypeName(e ast.Expr) string {
	switch t := e.(type) {
	case *ast.StarExpr:
		return baseTypeName(t.X)
	case *ast.IndexExpr:
		return baseTypeName(t.X)
	case *ast.IndexListExpr:
		return baseTypeName(t.X)
	case *ast.Ident:
		return t.Name
	}
	return ""
}
