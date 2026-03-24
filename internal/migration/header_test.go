package migration

import "testing"

func TestParseAndWriteHeader(t *testing.T) {
	s := WithHeader("# Title\n", "docs/architecture/x.md", "9")
	h, ok := ParseHeader(s)
	if !ok || h.Version != "9" || h.Template != "docs/architecture/x.md" {
		t.Fatalf("unexpected header: %+v ok=%v", h, ok)
	}
}

func TestWithHeaderReplacesExisting(t *testing.T) {
	in := "<!-- centinela:doc-version=1 template=a.md -->\n# Old\n"
	out := WithHeader(in, "b.md", "2")
	h, ok := ParseHeader(out)
	if !ok || h.Version != "2" || h.Template != "b.md" {
		t.Fatalf("unexpected header after replace: %+v ok=%v", h, ok)
	}
}
