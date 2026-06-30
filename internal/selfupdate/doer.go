// Package selfupdate implements `centinela update`: it resolves the latest
// GitHub release, downloads the host-platform asset, verifies it against the
// release SHA256SUMS, and atomically replaces the running binary. It also backs
// the throttled, fail-silent SessionStart "update available" notice. It is a
// pure leaf: it imports only the standard library and computes its XDG cache
// path directly from the environment, so every path is testable offline behind
// an injected Doer. It never auto-installs from the notice path.
package selfupdate

import (
	"net/http"
	"runtime"
	"time"
)

// Doer is the minimal HTTP seam; tests inject an httptest.Server-backed client
// so no real network or real GitHub is ever contacted.
type Doer interface {
	Do(*http.Request) (*http.Response, error)
}

const (
	defaultAPIBase = "https://api.github.com"
	repoSlug       = "samuelnp/centinela"
	devVersion     = "dev"
	sumsAsset      = "SHA256SUMS"
	defaultTTL     = 24 * time.Hour
	devMessage     = "centinela update: this is a development build — self-update is not available"
)

// Updater holds the injectable configuration for a self-update run. Production
// code builds it via New; tests construct a literal with their own Doer,
// Version, platform, clock, and Target resolver.
type Updater struct {
	Version string
	GOOS    string
	GOARCH  string
	APIBase string
	HTTP    Doer
	TTL     time.Duration
	Now     func() time.Time
	Target  func() (string, error)
}

// New returns an Updater wired to the real host platform, the real GitHub API,
// and the running binary version.
func New(version string) *Updater {
	return &Updater{
		Version: version,
		GOOS:    runtime.GOOS,
		GOARCH:  runtime.GOARCH,
		APIBase: defaultAPIBase,
		HTTP:    http.DefaultClient,
		TTL:     defaultTTL,
		Now:     time.Now,
		Target:  targetPath,
	}
}
