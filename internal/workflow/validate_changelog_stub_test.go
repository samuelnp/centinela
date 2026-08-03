package workflow

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// changelogStub is what `centinela artifact new <f> changelog` writes.
const changelogStub = "- <FILL: type>: <FILL: one-line summary of the change>"

// writeChangelog drops body at .workflow/<feature>-changelog.md in the fixture.
func writeChangelog(t *testing.T, feature, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(WorkflowDir, feature+"-changelog.md"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestValidateChangelogRejectsStubBehindPreamble pins the bypass that made the
// stub check cosmetic: it inspected only the FIRST non-blank line, so any
// heading or preamble re-opened the gate on an untouched scaffold. The rule is
// now content-scoped, not position-scoped — an unreplaced slot ANYWHERE fails,
// and a half-filled entry is still a template.
func TestValidateChangelogRejectsStubBehindPreamble(t *testing.T) {
	for name, body := range map[string]string{
		"bare stub":                 changelogStub + "\n",
		"markdown heading":          "## Changelog\n\n" + changelogStub + "\n",
		"prose preamble":            "Changelog:\n" + changelogStub + "\n",
		"title and prose":           "# in\n\nEntries for this feature:\n\n" + changelogStub + "\n",
		"type filled only":          "- fix: <FILL: one-line summary of the change>\n",
		"summary filled only":       "- <FILL: type>: tidy the docs step\n",
		"stub after a filled entry": "- feat: something real\n" + changelogStub + "\n",
		"crlf line endings":         "## Changelog\r\n\r\n" + changelogStub + "\r\n",
		"byte order mark":           "\ufeff" + changelogStub + "\n",
	} {
		t.Run(name, func(t *testing.T) {
			internalDocsFixture(t, "in")
			writeChangelog(t, "in", body)
			err := validateChangelog("in")
			if err == nil {
				t.Fatalf("unreplaced slot must fail the docs gate: %q", body)
			}
			if !strings.Contains(err.Error(), "template placeholder") {
				t.Fatalf("error must name the placeholder, got %v", err)
			}
		})
	}
}

// The rule must not fire on a changelog that legitimately QUOTES the marker.
// The generic citation form `<FILL: ...>` names no substance to replace, so it
// is a quotation rather than an unreplaced slot — and it is the single most
// natural line this very feature ships, which is what makes it a real case
// rather than a theoretical one.
func TestValidateChangelogAcceptsProseQuotingTheMarker(t *testing.T) {
	for name, body := range map[string]string{
		"this feature's own entry": "- fix: reject an unreplaced <FILL: ...> marker behind a preamble\n",
		"plain entry":              "- feat: bind the evidence gates so a stub can no longer pass\n",
		"quoted below a heading":   "## Changelog\n\n- fix: reject an unreplaced <FILL: ...> marker\n",
		"citation in later prose":  "- feat: real entry\n\nThe <FILL: ...> form is a citation, not a slot.\n",
		"unicode ellipsis":         "- fix: the <FILL: …> marker is now content-scoped\n",
		"unterminated marker":      "- docs: mention <FILL: without a closer\n",
	} {
		t.Run(name, func(t *testing.T) {
			internalDocsFixture(t, "in")
			writeChangelog(t, "in", body)
			if err := validateChangelog("in"); err != nil {
				t.Fatalf("legitimate changelog rejected (%q): %v", body, err)
			}
		})
	}
}

// The empty and missing cases keep their own distinct remedies: a blank file
// is not a template, and telling an author to "replace the slots" in a file
// that has none would send them looking for something that is not there.
func TestValidateChangelogKeepsEmptyAndMissingDistinct(t *testing.T) {
	internalDocsFixture(t, "in")
	if err := validateChangelog("in"); err == nil || !strings.Contains(err.Error(), "changelog entry missing") {
		t.Fatalf("missing changelog: got %v", err)
	}
	writeChangelog(t, "in", "   \n\t\n")
	if err := validateChangelog("in"); err == nil || !strings.Contains(err.Error(), "changelog entry is empty") {
		t.Fatalf("blank changelog: got %v", err)
	}
}
