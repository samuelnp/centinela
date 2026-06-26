package delivery

import "testing"

func TestComposeChangelogCategoryFromStub(t *testing.T) {
	cases := []struct{ stub, wantCat, wantLine string }{
		{"- feat: add thing", "Added", "feat: add thing"},
		{"- fix: a bug", "Fixed", "fix: a bug"},
		{"- refactor: tidy", "Changed", "refactor: tidy"},
		{"- chore: misc", "Changed", "chore: misc"},
	}
	for _, c := range cases {
		got := ComposeChangelog(Evidence{ChangelogStub: c.stub})
		if got.Category != c.wantCat || got.Line != c.wantLine {
			t.Fatalf("stub %q -> %+v want {%s %s}", c.stub, got, c.wantCat, c.wantLine)
		}
	}
}

func TestComposeChangelogFirstNonFillLine(t *testing.T) {
	stub := "# heading\n\nFILL: replace me\n- feat: real line\n"
	got := ComposeChangelog(Evidence{ChangelogStub: stub})
	if got.Line != "feat: real line" || got.Category != "Added" {
		t.Fatalf("FILL ignored, seed real line: %+v", got)
	}
}

func TestComposeChangelogDeriveFromBrief(t *testing.T) {
	stub := "FILL: <one line>\n"
	brief := "## Problem\n\nusers cannot log in\n\nmore\n"
	got := ComposeChangelog(Evidence{Feature: "alpha", ChangelogStub: stub, Brief: brief})
	if got.Line != "alpha: users cannot log in" || got.Category != "Changed" {
		t.Fatalf("derive from brief: %+v", got)
	}
}

func TestComposeChangelogFallbackToSlug(t *testing.T) {
	got := ComposeChangelog(Evidence{Feature: "alpha", ChangelogStub: "FILL\n"})
	if got.Line != "alpha" || got.Category != "Changed" {
		t.Fatalf("fallback to slug: %+v", got)
	}
}
