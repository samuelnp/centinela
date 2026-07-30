package orchestration

import (
	"os"
	"testing"
)

func surfaceBrief(t *testing.T, body string) {
	t.Helper()
	t.Chdir(t.TempDir())
	if err := os.MkdirAll("docs/features", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("docs/features/f.md", []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// Real briefs declare the surface bare, blockquoted, or as a list item — every
// marker form must classify user-facing (the gate discriminator).
func TestIsUserFacingFeatureSurfaceLineFormats(t *testing.T) {
	cases := map[string]string{
		"bare":            "# f\nsurface: user-facing\n",
		"bare mixed case": "# f\nSurface: User-Facing\n",
		"underscore":      "# f\nsurface: user_facing\n",
		"blockquote":      "# f\n> surface: user-facing\n",
		"list dash":       "# f\n- surface: user-facing\n",
		"list star":       "# f\n* surface: user-facing\n",
		"indented dash":   "# f\n  - surface: user-facing\n",
	}
	for name, body := range cases {
		surfaceBrief(t, body)
		if !IsUserFacingFeature("f") {
			t.Fatalf("%s: %q must classify user-facing", name, body)
		}
	}
}

// Marker stripping keeps the line-start discipline: prose mentioning
// "surface:" mid-line, internal values, and the bold form stay internal.
func TestIsUserFacingFeatureSurfaceLineNegatives(t *testing.T) {
	cases := map[string]string{
		"prose mid-line":  "# f\nthe user-facing surface: is discussed here\n",
		"list prose":      "# f\n- the surface: user-facing idea\n",
		"internal bare":   "# f\nsurface: internal\n",
		"internal list":   "# f\n- surface: internal\n",
		"internal quote":  "# f\n> surface: internal\n",
		"bold form":       "# f\n**Surface:** user-facing\n",
		"no surface line": "# f\nnothing declared\n",
	}
	for name, body := range cases {
		surfaceBrief(t, body)
		if IsUserFacingFeature("f") {
			t.Fatalf("%s: %q must classify internal", name, body)
		}
	}
}
