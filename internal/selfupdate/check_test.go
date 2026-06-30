package selfupdate

import (
	"strings"
	"testing"
	"time"
)

func TestCheckBehind(t *testing.T) {
	fs := newServer(t, fakeRelease{tag: "v0.40.2"})
	u, _ := newUpdater(t, fs, "0.37.0")
	res, err := u.Check()
	if err != nil {
		t.Fatal(err)
	}
	if !res.Behind || !strings.Contains(res.Message, "v0.40.2") {
		t.Fatalf("res = %+v", res)
	}
}

func TestCheckCurrent(t *testing.T) {
	fs := newServer(t, fakeRelease{tag: "v0.40.2"})
	u, _ := newUpdater(t, fs, "0.40.2")
	res, err := u.Check()
	if err != nil {
		t.Fatal(err)
	}
	if res.Behind || !strings.Contains(res.Message, "up to date") {
		t.Fatalf("res = %+v", res)
	}
}

func TestCheckHonorsCacheNoNetwork(t *testing.T) {
	fs := newServer(t, fakeRelease{tag: "v0.40.2"})
	u, _ := newUpdater(t, fs, "0.37.0")
	writeCacheFile(t, u, "v0.40.2", time.Hour)
	res, err := u.Check()
	if err != nil {
		t.Fatal(err)
	}
	if !res.Behind || fs.hits != 0 {
		t.Fatalf("expected cache hit, behind=%v hits=%d", res.Behind, fs.hits)
	}
}

func TestCheckDevBuild(t *testing.T) {
	fs := newServer(t, fakeRelease{tag: "v0.40.2"})
	u, _ := newUpdater(t, fs, "dev")
	res, err := u.Check()
	if err != nil || res.Behind || fs.hits != 0 {
		t.Fatalf("dev check res=%+v err=%v hits=%d", res, err, fs.hits)
	}
}

func TestCheckNetworkError(t *testing.T) {
	u, _ := newUpdater(t, nil, "0.37.0")
	u.HTTP = errDoer{}
	u.APIBase = "http://example.invalid"
	if _, err := u.Check(); err == nil {
		t.Fatal("want network error")
	}
}
