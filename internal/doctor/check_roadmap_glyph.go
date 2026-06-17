package doctor

import (
	"strings"
	"unicode"

	"github.com/samuelnp/centinela/internal/roadmap"
)

// glyphPhases returns the names of phases whose name begins with a leading
// non-letter status glyph before "Phase" (e.g. "✅ Phase 0: Bootstrap"). A
// leading glyph defeats the lowercase "phase 0" prefix match used by bootstrap
// detection, silently blocking greenfield starts.
func glyphPhases(rm *roadmap.Roadmap) []string {
	var out []string
	for _, p := range rm.Phases {
		if hasLeadingGlyph(p.Name) {
			out = append(out, p.Name)
		}
	}
	return out
}

// hasLeadingGlyph reports whether name has a leading non-letter rune followed
// (after trimming leading glyphs+spaces) by the word "Phase". This targets a
// status glyph prefix, not arbitrary punctuation in the middle of a name.
func hasLeadingGlyph(name string) bool {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return false
	}
	first := []rune(trimmed)[0]
	if unicode.IsLetter(first) {
		return false
	}
	return strings.HasPrefix(strings.ToLower(stripLeadingGlyph(name)), "phase")
}

// stripLeadingGlyph removes a leading run of non-letter, non-digit runes (the
// status glyph) and the whitespace after it. "✅ Phase 0: Bootstrap" becomes
// "Phase 0: Bootstrap"; a clean name is returned unchanged.
func stripLeadingGlyph(name string) string {
	runes := []rune(name)
	i := 0
	for i < len(runes) && !unicode.IsLetter(runes[i]) && !unicode.IsDigit(runes[i]) {
		i++
	}
	return strings.TrimSpace(string(runes[i:]))
}

// repairRoadmap strips any leading phase-name glyphs and regenerates ROADMAP.md
// from the (possibly repaired) roadmap.json. roadmap.json is rewritten only
// when a glyph was actually stripped, so a drift-only repair never reformats
// it. Idempotent: a clean roadmap yields byte-identical roadmap.json and
// ROADMAP.md on re-run.
func repairRoadmap() error {
	rm, err := roadmap.Load()
	if err != nil {
		return err
	}
	stripped := false
	for i := range rm.Phases {
		clean := stripLeadingGlyph(rm.Phases[i].Name)
		if clean != rm.Phases[i].Name {
			rm.Phases[i].Name = clean
			stripped = true
		}
	}
	if stripped {
		if err := roadmap.Save(rm); err != nil {
			return err
		}
	}
	return writeRoadmapMarkdown(rm)
}
