package migration

import "testing"

func TestParseHeaderInvalid(t *testing.T) {
	if _, ok := ParseHeader("# title\n"); ok {
		t.Fatal("expected invalid header parse")
	}
	if _, ok := ParseHeader("<!-- centinela:doc-version= template=x -->\n"); ok {
		t.Fatal("expected parse failure for empty version")
	}
}

func TestWithHeaderEmptyBody(t *testing.T) {
	out := WithHeader("", "docs/architecture/x.md", "1")
	h, ok := ParseHeader(out)
	if !ok || h.Template != "docs/architecture/x.md" {
		t.Fatal("expected header in empty output")
	}
}
