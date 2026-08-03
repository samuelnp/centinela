package gates

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/gitdiff"
)

// F2 regression, on the GATE path: go/doc treats a trailing line comment as
// documentation and godoc renders it, so rejecting it fails the build for text
// the pipeline this gate feeds would publish.
func TestCheckDocstring_TrailingLineCommentIsDocumentation(t *testing.T) {
	inDir(t)
	src := "package a\n\nconst Trailing = 1 // Trailing is the answer.\n\n" +
		"var VarTrailing = 2 // VarTrailing is a var.\n\n" +
		"// S is a struct.\ntype S struct{}\n"
	p := fixture(t, "a.go", src)
	r := checkDocstring(docstringCfg("fail"), gitdiff.NewSet([]string{p}))
	if r.Status != Pass {
		t.Fatalf("status = %v (%s) details=%v", r.Status, r.Message, r.Details)
	}
}

// A trailing group holding only the directive is still undocumented, so it
// falls through to the exemption branch rather than silently counting as docs.
func TestCheckDocstring_TrailingNodocIsExemptNotDocumented(t *testing.T) {
	inDir(t)
	p := fixture(t, "a.go", "package a\n\nvar Exported = 1 //centinela:nodoc\n")
	r := checkDocstring(docstringCfg("fail"), gitdiff.NewSet([]string{p}))
	if r.Status != Pass {
		t.Fatalf("status = %v (%s)", r.Status, r.Message)
	}
	if !strings.Contains(r.Message, "exempt via //centinela:nodoc") {
		t.Fatalf("exemption must not be reported as documented: %q", r.Message)
	}
}

// Trailing comments obey require_name_prefix exactly like doc comments do.
func TestCheckDocstring_TrailingCommentHonorsRequireNamePrefix(t *testing.T) {
	inDir(t)
	p := fixture(t, "a.go", "package a\n\nconst Trailing = 1 // wrongly named.\n")
	cfg := docstringCfg("fail")
	if r := checkDocstring(cfg, gitdiff.NewSet([]string{p})); r.Status != Pass {
		t.Fatalf("prefix off must pass: %v %v", r.Status, r.Details)
	}
	cfg.Gates.Docstring.RequireNamePrefix = true
	if r := checkDocstring(cfg, gitdiff.NewSet([]string{p})); r.Status != Fail {
		t.Fatalf("prefix on must flag the trailing comment: %v", r.Status)
	}
}
