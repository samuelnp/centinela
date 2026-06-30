package acceptance_test

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"
	"time"

	"github.com/samuelnp/centinela/internal/selfupdate"
)

func newAcUpdater(t *testing.T, version string, o acFakeOpts) (*acSrv, *selfupdate.Updater, string) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("XDG_CACHE_HOME", filepath.Join(dir, "cache"))
	bin := filepath.Join(dir, "centinela")
	if err := os.WriteFile(bin, []byte("OLD"), 0o755); err != nil {
		t.Fatal(err)
	}
	srv := newAcServer(t, o)
	u := &selfupdate.Updater{Version: version, GOOS: runtime.GOOS, GOARCH: runtime.GOARCH,
		APIBase: srv.URL, HTTP: srv.Client(), TTL: 24 * time.Hour, Now: time.Now,
		Target: func() (string, error) { return bin, nil }}
	return srv, u, bin
}

func acCachePath() string {
	xdg := os.Getenv("XDG_CACHE_HOME")
	if xdg == "" {
		xdg = filepath.Join(os.Getenv("HOME"), ".cache")
	}
	return filepath.Join(xdg, "centinela", "update-check.json")
}

func seedCache(t *testing.T, u *selfupdate.Updater, tag string, age time.Duration) {
	t.Helper()
	ts := time.Unix(1_700_000_000, 0)
	u.Now = func() time.Time { return ts }
	at := ts.Add(-age).Unix()
	p := acCachePath()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	body := `{"latestTag":"` + tag + `","checkedAt":` + strconv.FormatInt(at, 10) + `}`
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// acErrDoer simulates an offline network for deterministic error tests.
type acErrDoer struct{}

func (acErrDoer) Do(*http.Request) (*http.Response, error) { return nil, errors.New("offline") }
