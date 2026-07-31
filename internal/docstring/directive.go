package docstring

import (
	"go/ast"
	"regexp"
	"strings"
)

// generatedRe matches Go's canonical generated-file banner.
var generatedRe = regexp.MustCompile(`^// Code generated .* DO NOT EDIT\.$`)

// isGenerated reports whether a comment before the package clause carries the
// canonical generated-file banner. Generated files are never inspected.
func isGenerated(f *ast.File) bool {
	for _, g := range f.Comments {
		if g.Pos() > f.Package {
			return false
		}
		for _, c := range g.List {
			if generatedRe.MatchString(strings.TrimRight(c.Text, "\r")) {
				return true
			}
		}
	}
	return false
}

// hasNodoc reports whether a comment group carries the opt-out Directive.
func hasNodoc(groups ...*ast.CommentGroup) bool {
	for _, g := range groups {
		if g == nil {
			continue
		}
		for _, c := range g.List {
			if strings.TrimSpace(c.Text) == Directive {
				return true
			}
		}
	}
	return false
}

// documented reports whether a doc comment counts as documentation. v1 asks
// only for non-empty text; requirePrefix additionally demands the golint-style
// "Name ..." opening. Note ast.CommentGroup.Text strips directive comments, so
// a group holding only the nodoc Directive is correctly not documentation.
func documented(g *ast.CommentGroup, name string, requirePrefix bool) bool {
	if g == nil {
		return false
	}
	text := strings.TrimSpace(g.Text())
	if text == "" {
		return false
	}
	if requirePrefix {
		return text == name || strings.HasPrefix(text, name+" ")
	}
	return true
}

// firstDoc returns the spec's own doc comment, falling back to the enclosing
// GenDecl's — a doc on `const (` covers every constant in the block, which is
// how godoc renders block-documented declarations.
func firstDoc(own, block *ast.CommentGroup) *ast.CommentGroup {
	if own != nil {
		return own
	}
	return block
}
