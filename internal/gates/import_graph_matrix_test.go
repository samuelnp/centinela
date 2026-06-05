package gates

import (
	"testing"

	"github.com/samuelnp/centinela/internal/config"
)

func sampleLayers() []config.Layer {
	return []config.Layer{
		{Name: "leaf", Paths: []string{"internal/config/**"}, Allow: nil},
		{Name: "domain", Paths: []string{"internal/gates/**", "internal/workflow/**"}, Allow: []string{"leaf"}},
		{Name: "cmd", Paths: []string{"cmd/**"}, Allow: []string{"domain", "leaf"}},
	}
}

func TestBuildMatrix_ValidatesAndUnionsDuplicates(t *testing.T) {
	layers := append(sampleLayers(),
		config.Layer{Name: "domain", Paths: []string{"internal/extra/**"}, Allow: []string{"leaf"}})
	m, err := buildMatrix(layers)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !m.allow["domain"]["leaf"] || !m.allowed("cmd", "domain") {
		t.Fatalf("allow-sets wrong: %+v", m.allow)
	}
	if m.layerFor("internal/extra/x") != "domain" {
		t.Fatal("duplicate layer paths should union")
	}
}

func TestBuildMatrix_Errors(t *testing.T) {
	cases := []struct {
		name   string
		layers []config.Layer
		want   string
	}{
		{"empty name", []config.Layer{{Name: "", Paths: []string{"a/**"}}}, "empty name"},
		{"no paths", []config.Layer{{Name: "x", Paths: nil}}, `layer "x" has no paths`},
		{"unknown allow", []config.Layer{{Name: "x", Paths: []string{"a/**"}, Allow: []string{"ghost"}}}, `unknown layer "ghost"`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := buildMatrix(tc.layers)
			if err == nil || !contains(err.Error(), tc.want) {
				t.Fatalf("want error containing %q, got %v", tc.want, err)
			}
		})
	}
}

func TestLayerForAndGlobMatch(t *testing.T) {
	m, _ := buildMatrix(sampleLayers())
	cases := map[string]string{
		"internal/config":        "leaf",
		"internal/config/sub/x":  "leaf",
		"internal/gates/foo":     "domain",
		"cmd/centinela":          "cmd",
		"internal/configx/thing": "", // segment-boundary: not under internal/config
		"internal/ui/render":     "", // unmapped
	}
	for rel, want := range cases {
		if got := m.layerFor(rel); got != want {
			t.Errorf("layerFor(%q)=%q want %q", rel, got, want)
		}
	}
}

func TestGlobMatch_PathMatchSemantics(t *testing.T) {
	if !globMatch("cmd/*", "cmd/main") {
		t.Fatal("single-segment glob should match")
	}
	if globMatch("cmd/*", "cmd/a/b") {
		t.Fatal("single-* must not cross a segment")
	}
	if !globMatch("**", "anything/here") {
		t.Fatal("bare ** must match the whole tree")
	}
}
