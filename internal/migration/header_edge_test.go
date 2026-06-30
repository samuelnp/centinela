package migration

import (
	"strings"
	"testing"
)

func TestParseHeaderSingleToken(t *testing.T) {
	// Only one token (no template= field) -> len(toks) < 2 rejects the header.
	if _, ok := ParseHeader("<!-- centinela:doc-version=1 -->\n"); ok {
		t.Fatal("expected single-token header to be rejected")
	}
}

func TestParseHeaderEmptyTemplate(t *testing.T) {
	// Present but empty template= value -> rejected by the v=="" || t=="" guard.
	if _, ok := ParseHeader("<!-- centinela:doc-version=1 template= -->\n"); ok {
		t.Fatal("expected empty template value to be rejected")
	}
}

func TestWithHeaderExistingNoNewline(t *testing.T) {
	// A valid header with no body and no trailing newline: strings.Cut finds no
	// newline (found==false), so the existing header is dropped to empty content.
	in := "<!-- centinela:doc-version=1 template=a.md -->"
	out := WithHeader(in, "b.md", "2")
	h, ok := ParseHeader(out)
	if !ok || h.Version != "2" || h.Template != "b.md" {
		t.Fatalf("unexpected header after replace: %+v ok=%v", h, ok)
	}
	if strings.Count(out, "\n") != 1 {
		t.Fatalf("expected single trailing newline, got %q", out)
	}
}
