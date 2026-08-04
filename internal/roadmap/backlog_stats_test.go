package roadmap

import "testing"

func TestSummarizeBacklogCounts(t *testing.T) {
	r := backlogOf(spec{"old", daysAgo(40)}, spec{"mid", daysAgo(20)}, spec{"young", daysAgo(2)})
	stats := SummarizeBacklog(AgeBacklog(r, ageNow, DefaultStaleDays), DefaultStaleDays)
	if stats.Total != 3 || stats.Stale != 2 || stats.ThresholdDays != DefaultStaleDays {
		t.Fatalf("stats = %+v", stats)
	}
	if !stats.OldestKnown || stats.OldestDays != 40 || stats.OldestSlug != "old" {
		t.Fatalf("oldest = %+v", stats)
	}
}

// An all-unknown Backlog has no oldest to name, but is fully stale.
func TestSummarizeBacklogWithNoParseableClock(t *testing.T) {
	r := backlogOf(spec{"a", ""}, spec{"b", "garbage"})
	stats := SummarizeBacklog(AgeBacklog(r, ageNow, 0), 0)
	if stats.Total != 2 || stats.Stale != 2 {
		t.Fatalf("stats = %+v", stats)
	}
	if stats.OldestKnown || stats.OldestSlug != "" {
		t.Fatalf("no oldest may be claimed: %+v", stats)
	}
	if stats.ThresholdDays != DefaultStaleDays {
		t.Fatalf("a non-positive threshold must fall back to the default: %+v", stats)
	}
}

func TestSummarizeBacklogEmpty(t *testing.T) {
	stats := SummarizeBacklog(nil, 30)
	if stats.Total != 0 || stats.Stale != 0 || stats.OldestKnown {
		t.Fatalf("stats = %+v", stats)
	}
	if stats.ThresholdDays != 30 {
		t.Fatalf("threshold must be carried through: %+v", stats)
	}
}
