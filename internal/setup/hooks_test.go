package setup

import "testing"

func TestMergeHooksAndIdempotency(t *testing.T) {
	var pre, post, prompt []HookGroup
	if !mergeHooks(&pre, &post, &prompt) {
		t.Fatal("expected first merge to change")
	}
	if len(pre) != 2 || len(post) != 2 || len(prompt) != 7 {
		t.Fatalf("unexpected group sizes: %d %d %d", len(pre), len(post), len(prompt))
	}
	if mergeHooks(&pre, &post, &prompt) {
		t.Fatal("expected second merge to be no-op")
	}
}

func TestEnsureGroupAndPrompt(t *testing.T) {
	groups := []HookGroup{}
	if !ensureGroup(&groups, "Write", "x", "m") {
		t.Fatal("ensureGroup should add")
	}
	if ensureGroup(&groups, "Write", "x", "m") {
		t.Fatal("ensureGroup should not duplicate")
	}
	p := []HookGroup{}
	if !ensurePrompt(&p, "cmd", "msg") {
		t.Fatal("ensurePrompt should add")
	}
	if ensurePrompt(&p, "cmd", "msg") {
		t.Fatal("ensurePrompt should not duplicate")
	}
}
