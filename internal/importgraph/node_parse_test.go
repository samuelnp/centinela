package importgraph

import "testing"

const depcruiseFixture = `{"modules":[
 {"source":"./src/a.js","dependencies":[
   {"resolved":"src/b.js","coreModule":false},
   {"resolved":"fs","coreModule":true},
   {"resolved":"node_modules/lodash/index.js","coreModule":false},
   {"resolved":"src/b.js","coreModule":false}]},
 {"source":"src/b.js","dependencies":[]},
 {"source":"node_modules/x/i.js","dependencies":[]}]}`

func TestParseDepcruise(t *testing.T) {
	pkgs, err := parseDepcruise([]byte(depcruiseFixture))
	if err != nil {
		t.Fatal(err)
	}
	if len(pkgs) != 2 {
		t.Fatalf("node_modules source must be dropped; got %+v", pkgs)
	}
	a := pkgs[0]
	if a.Path != "src/a.js" || len(a.Imports) != 1 || a.Imports[0] != "src/b.js" {
		t.Fatalf("core/node_modules/dupe deps must be dropped: %+v", a)
	}
}

func TestParseDepcruise_Malformed(t *testing.T) {
	if _, err := parseDepcruise([]byte("{not json")); err == nil {
		t.Fatal("malformed JSON must error, never a silent empty graph")
	}
}

func TestParseMadge(t *testing.T) {
	out := `{"src/a.js":["src/b.js","node_modules/x/i.js","src/a.js"],"src/b.js":[]}`
	pkgs, err := parseMadge([]byte(out))
	if err != nil {
		t.Fatal(err)
	}
	if len(pkgs) != 2 || pkgs[0].Path != "src/a.js" || pkgs[1].Path != "src/b.js" {
		t.Fatalf("madge output must be sorted by path: %+v", pkgs)
	}
	if len(pkgs[0].Imports) != 1 || pkgs[0].Imports[0] != "src/b.js" {
		t.Fatalf("node_modules + self deps must be dropped: %+v", pkgs[0])
	}
}

func TestParseMadge_Malformed(t *testing.T) {
	if _, err := parseMadge([]byte("[]")); err == nil {
		t.Fatal("non-object JSON must error")
	}
}

func TestNormalizeNodePath(t *testing.T) {
	cases := map[string]string{
		"./src/a.js":              "src/a.js",
		"src/b.js":                "src/b.js",
		"node_modules/x":          "",
		"pkg/node_modules/y/i.js": "",
		"":                        "",
	}
	for in, want := range cases {
		if got := normalizeNodePath(in); got != want {
			t.Errorf("normalizeNodePath(%q)=%q want %q", in, got, want)
		}
	}
}
