package roadmap

import (
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

// The end-to-end proof: mutations entering through BOTH spellings concurrently
// must all survive. Before canonicalization this destroyed 6 of 24 records and
// printed a green "✓ Committed" for one that was gone.
func TestConcurrentDefersAcrossASymlinkedPathAllSurvive(t *testing.T) {
	realPath, linkPath := symlinkRepo(t)
	const perSide = 4
	var wg sync.WaitGroup
	errs := make([]error, 2*perSide)
	start := make(chan struct{})
	for i := 0; i < 2*perSide; i++ {
		path, slug := realPath, "via-real-"+string(rune('a'+i))
		if i >= perSide {
			path, slug = linkPath, "via-link-"+string(rune('a'+i))
		}
		wg.Add(1)
		go func(i int, path, slug string) {
			defer wg.Done()
			<-start
			errs[i] = Defer(path, DeferOptions{Slug: slug, Summary: "s",
				Now: time.Date(2026, 8, 4, 12, 0, i, 0, time.UTC)})
		}(i, path, slug)
	}
	close(start)
	wg.Wait()
	body, err := os.ReadFile(realPath)
	if err != nil {
		t.Fatal(err)
	}
	for i, err := range errs {
		if err != nil {
			t.Fatalf("defer %d: %v", i, err)
		}
	}
	if got := strings.Count(string(body), `"deferredAt"`); got != 2*perSide {
		t.Fatalf("%d of %d records survived — the symlinked spelling re-opened the race:\n%s",
			got, 2*perSide, body)
	}
}
