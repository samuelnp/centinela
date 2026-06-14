package doctor

import "testing"

func TestHasLeadingGlyph(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want bool
	}{
		{"clean", "Phase 0: Bootstrap", false},
		{"leading-glyph", "✅ Phase 0: Bootstrap", true},
		{"digit-leading", "0 Phase", false},
		{"empty", "", false},
		{"whitespace", "   ", false},
		{"glyph-no-phase", "✅ Milestone", false},
		{"midword-punct", "Phase 1: A-B", false},
	}
	for _, c := range cases {
		if got := hasLeadingGlyph(c.in); got != c.want {
			t.Errorf("%s: hasLeadingGlyph(%q)=%v want %v", c.name, c.in, got, c.want)
		}
	}
}

func TestStripLeadingGlyph(t *testing.T) {
	cases := []struct{ in, want string }{
		{"Phase 0: Bootstrap", "Phase 0: Bootstrap"},
		{"✅ Phase 0: Bootstrap", "Phase 0: Bootstrap"},
		{"0 Phase", "0 Phase"},
		{"", ""},
		{"✅   Phase 0", "Phase 0"},
		{"✅✨ Phase 0", "Phase 0"},
	}
	for _, c := range cases {
		if got := stripLeadingGlyph(c.in); got != c.want {
			t.Errorf("stripLeadingGlyph(%q)=%q want %q", c.in, got, c.want)
		}
	}
}

func TestDescribeRoadmap(t *testing.T) {
	if m := describeRoadmap([]string{"✅ Phase 0"}, true); m == "" {
		t.Fatal("glyph+drift message empty")
	}
	if m := describeRoadmap([]string{"✅ Phase 0"}, false); m == "" {
		t.Fatal("glyph-only message empty")
	}
	if m := describeRoadmap(nil, true); m == "" {
		t.Fatal("drift-only message empty")
	}
}
