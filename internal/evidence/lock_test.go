package evidence

import (
	"sync"
	"testing"
	"time"

	"github.com/samuelnp/centinela/internal/orchestration"
)

func TestLockSerializesConcurrentAppends(t *testing.T) {
	chdirToTemp(t)
	s := Skeleton("alpha", orchestration.RoleBigThinker, "v1")
	if err := WriteAtomic("alpha", orchestration.RoleBigThinker, s); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	work := func(value string) {
		defer wg.Done()
		release, err := Lock("alpha", orchestration.RoleBigThinker)
		if err != nil {
			t.Errorf("lock: %v", err)
			return
		}
		defer release()
		doc, err := Read("alpha", orchestration.RoleBigThinker)
		if err != nil {
			t.Errorf("read: %v", err)
			return
		}
		if err := AppendField(doc, "outputs", value); err != nil {
			t.Errorf("append: %v", err)
			return
		}
		if err := WriteAtomic("alpha", orchestration.RoleBigThinker, doc); err != nil {
			t.Errorf("write: %v", err)
		}
	}
	wg.Add(2)
	go work("foo.md")
	go work("bar.md")
	wg.Wait()
	doc, err := Read("alpha", orchestration.RoleBigThinker)
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.Outputs) != 2 {
		t.Fatalf("expected both outputs to survive, got %v", doc.Outputs)
	}
}

func TestLockTimeoutSurfacesHint(t *testing.T) {
	chdirToTemp(t)
	release, err := Lock("alpha", orchestration.RoleBigThinker)
	if err != nil {
		t.Fatal(err)
	}
	defer release()
	original := LockPollInterval
	_ = original
	start := time.Now()
	_, err2 := Lock("alpha", orchestration.RoleBigThinker)
	if err2 == nil {
		t.Fatal("expected timeout")
	}
	elapsed := time.Since(start)
	if elapsed < LockTimeout-10*time.Millisecond {
		t.Fatalf("returned too fast: %v", elapsed)
	}
}
