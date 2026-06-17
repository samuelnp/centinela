package githooks

import (
	"strings"
	"testing"
)

func TestSplice_AppendsIntoEmpty(t *testing.T) {
	out, changed := splice("", Block)
	if !changed {
		t.Fatal("splice into empty must report changed")
	}
	if out != Block {
		t.Fatalf("empty splice must equal the bare block, got %q", out)
	}
}

func TestSplice_IdempotentReSplice(t *testing.T) {
	once, _ := splice("", Block)
	twice, changed := splice(once, Block)
	if changed {
		t.Fatal("re-splicing the identical block must report changed=false")
	}
	if twice != once {
		t.Fatalf("idempotent splice must be byte-identical:\n%q\n%q", once, twice)
	}
}

func TestSplice_PreservesUserContent(t *testing.T) {
	existing := "#!/bin/sh\necho pre-existing-hook\n"
	out, changed := splice(existing, Block)
	if !changed {
		t.Fatal("splice over user content must change the file")
	}
	if !strings.Contains(out, "echo pre-existing-hook") {
		t.Fatalf("user line not preserved: %q", out)
	}
	if !strings.Contains(out, BeginMarker) || !strings.Contains(out, "centinela precommit") {
		t.Fatalf("block not present after splice: %q", out)
	}
}

func TestSplice_ReplacesMarkedRegionInPlace(t *testing.T) {
	existing := "echo top\n" + BeginMarker + "\nOLD\n" + EndMarker + "\necho bottom\n"
	out, changed := splice(existing, Block)
	if !changed {
		t.Fatal("replacing a stale block must report changed")
	}
	if strings.Contains(out, "OLD") {
		t.Fatalf("stale block body must be replaced: %q", out)
	}
	if !strings.Contains(out, "echo top") || !strings.Contains(out, "echo bottom") {
		t.Fatalf("surrounding user lines must survive: %q", out)
	}
}

func TestRemoveBlock_RemovesOnlyMarkedRegion(t *testing.T) {
	existing := "#!/bin/sh\necho keep\n\n" + Block
	out, changed := removeBlock(existing)
	if !changed {
		t.Fatal("removeBlock must report changed when a block is present")
	}
	if strings.Contains(out, BeginMarker) || strings.Contains(out, "centinela precommit") {
		t.Fatalf("marked region must be gone: %q", out)
	}
	if !strings.Contains(out, "echo keep") {
		t.Fatalf("user line must survive removal: %q", out)
	}
}

func TestRemoveBlock_NoMarkersIsNoOp(t *testing.T) {
	existing := "#!/bin/sh\necho keep\n"
	out, changed := removeBlock(existing)
	if changed || out != existing {
		t.Fatalf("removeBlock with no markers must be a no-op, got changed=%v %q", changed, out)
	}
}
