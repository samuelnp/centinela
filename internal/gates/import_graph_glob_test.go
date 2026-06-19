package gates

import "testing"

func TestTrimDoubleStar(t *testing.T) {
	cases := []struct {
		in       string
		wantBase string
		wantOK   bool
	}{
		{"**", "", true},
		{"internal/config/**", "internal/config", true},
		{"cmd/*", "", false},
		{"internal/config", "", false},
	}
	for _, tc := range cases {
		base, ok := trimDoubleStar(tc.in)
		if base != tc.wantBase || ok != tc.wantOK {
			t.Errorf("trimDoubleStar(%q)=(%q,%v) want (%q,%v)", tc.in, base, ok, tc.wantBase, tc.wantOK)
		}
	}
}

func TestHasPrefixDir(t *testing.T) {
	if !hasPrefixDir("internal/config/x", "internal/config") {
		t.Fatal("should match a nested path")
	}
	if hasPrefixDir("internal/configx", "internal/config") {
		t.Fatal("must respect segment boundary")
	}
	if !hasPrefixDir("anything", "") {
		t.Fatal("empty base matches everything")
	}
}

func TestAllowed_SameLayerAndAllowList(t *testing.T) {
	m, _ := buildMatrix(sampleLayers())
	if !m.allowed("domain", "domain") {
		t.Fatal("same-layer imports must be allowed")
	}
	if !m.allowed("cmd", "leaf") {
		t.Fatal("allow-listed import must be permitted")
	}
	if m.allowed("leaf", "domain") {
		t.Fatal("non-allow-listed cross-layer import must be forbidden")
	}
}
