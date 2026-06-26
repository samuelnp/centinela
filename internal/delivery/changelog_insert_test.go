package delivery

import (
	"strings"
	"testing"
)

const clFull = "# Changelog\n\n## [Unreleased]\n\n### Added\n\n- existing add\n\n### Changed\n\n### Fixed\n\n---\n\n## [1.0.0]\n\n### Added\n\n- released thing\n"

func TestInsertEntryFirstThenIdempotent(t *testing.T) {
	out, changed := InsertEntry(clFull, ChangelogEntry{Category: "Added", Line: "feat: new"})
	if !changed || !strings.Contains(out, "- feat: new") {
		t.Fatalf("first insert should change: %v\n%s", changed, out)
	}
	out2, changed2 := InsertEntry(out, ChangelogEntry{Category: "Added", Line: "feat: new"})
	if changed2 || out2 != out {
		t.Fatalf("re-insert must be a no-op")
	}
	if strings.Count(out, "- feat: new") != 1 {
		t.Fatalf("exactly one copy expected:\n%s", out)
	}
}

func TestInsertEntryCreatesSubsectionInOrder(t *testing.T) {
	// No ### Added present; insert Added must land before Changed/Fixed.
	src := "## [Unreleased]\n\n### Changed\n\n- c1\n\n### Fixed\n\n- f1\n"
	out, changed := InsertEntry(src, ChangelogEntry{Category: "Added", Line: "feat: a"})
	if !changed {
		t.Fatal("should change")
	}
	ia := strings.Index(out, "### Added")
	ic := strings.Index(out, "### Changed")
	if ia < 0 || ia > ic {
		t.Fatalf("Added must precede Changed:\n%s", out)
	}
}

func TestInsertEntryReleasedSectionsUntouched(t *testing.T) {
	out, _ := InsertEntry(clFull, ChangelogEntry{Category: "Added", Line: "feat: new"})
	rel := out[strings.Index(out, "---"):]
	if strings.Count(rel, "- ") != 1 || !strings.Contains(rel, "released thing") {
		t.Fatalf("released section must be untouched:\n%s", rel)
	}
}

func TestInsertEntryNoUnreleasedBlock(t *testing.T) {
	src := "# Changelog\n\n## [1.0.0]\n\n### Added\n\n- x\n"
	out, changed := InsertEntry(src, ChangelogEntry{Category: "Added", Line: "feat: new"})
	if changed || out != src {
		t.Fatalf("no [Unreleased] -> unchanged+false")
	}
}

func TestInsertEntryNewFixedAfterAdded(t *testing.T) {
	// Only Added exists; new Fixed must be placed AFTER it (rank not > so falls
	// through to end), exercising newSubsectionAt's fall-through + unknown name.
	src := "## [Unreleased]\n\n### Added\n\n- a1\n\n### Misc\n\n- m1\n"
	out, changed := InsertEntry(src, ChangelogEntry{Category: "Fixed", Line: "fix: f"})
	if !changed || !strings.Contains(out, "### Fixed") {
		t.Fatalf("new Fixed subsection:\n%s", out)
	}
	if strings.Index(out, "### Added") > strings.Index(out, "### Fixed") {
		t.Fatalf("Fixed must come after Added:\n%s", out)
	}
}

func TestInsertEntrySubsectionAtEOF(t *testing.T) {
	// [Unreleased] runs to EOF with an existing Fixed; new Fixed bullet appends.
	src := "## [Unreleased]\n\n### Fixed\n\n- f1"
	out, changed := InsertEntry(src, ChangelogEntry{Category: "Fixed", Line: "fix: f2"})
	if !changed || !strings.Contains(out, "- fix: f2") {
		t.Fatalf("append at EOF:\n%s", out)
	}
}
