package insights

import (
	"testing"

	"github.com/samuelnp/centinela/internal/telemetry"
)

func blk(reason, fileType string) telemetry.Event {
	return telemetry.Event{Type: telemetry.TypeBlock, Reason: reason, FileType: fileType}
}

// blocks buckets by "<reason> · <fileType>", count desc then key asc.
func TestBlocksRanksByCountDescThenKeyAsc(t *testing.T) {
	ev := []telemetry.Event{
		blk("out-of-step", "plan"), blk("out-of-step", "plan"), blk("out-of-step", "plan"),
		blk("need-init", "source"), blk("need-init", "source"),
	}
	got := blocks(ev, 5)
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	if got[0].Key != "out-of-step · plan" || got[0].Count != 3 {
		t.Fatalf("first = %+v", got[0])
	}
	if got[1].Key != "need-init · source" || got[1].Count != 2 {
		t.Fatalf("second = %+v", got[1])
	}
}

// Equal counts break ties by key ascending.
func TestBlocksTieBreakKeyAsc(t *testing.T) {
	got := blocks([]telemetry.Event{blk("beta", "plan"), blk("alpha", "plan")}, 5)
	if got[0].Key != "alpha · plan" || got[1].Key != "beta · plan" {
		t.Fatalf("order = %+v", got)
	}
}

// An empty fileType buckets under reason + the <none> token.
func TestBlocksEmptyFileTypeBucketsAsNone(t *testing.T) {
	got := blocks([]telemetry.Event{blk("out-of-step", "")}, 5)
	if len(got) != 1 || got[0].Key != "out-of-step · <none>" {
		t.Fatalf("key = %+v", got)
	}
}

// Non-block events are ignored entirely.
func TestBlocksIgnoresOtherTypes(t *testing.T) {
	ev := []telemetry.Event{{Type: telemetry.TypeGateFailure}, blk("r", "f")}
	if got := blocks(ev, 5); len(got) != 1 || got[0].Count != 1 {
		t.Fatalf("got = %+v", got)
	}
}

// topN truncates the block section.
func TestBlocksTopNTruncates(t *testing.T) {
	ev := []telemetry.Event{blk("a", "x"), blk("b", "x"), blk("c", "x")}
	if got := blocks(ev, 2); len(got) != 2 {
		t.Fatalf("top 2 = %d entries", len(got))
	}
}
