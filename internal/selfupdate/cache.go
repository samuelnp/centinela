package selfupdate

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type cacheEntry struct {
	LatestTag string `json:"latestTag"`
	CheckedAt int64  `json:"checkedAt"`
}

// cachePath computes ${XDG_CACHE_HOME:-~/.cache}/centinela/update-check.json
// directly from the environment, keeping this package a pure leaf.
func cachePath() string {
	base := os.Getenv("XDG_CACHE_HOME")
	if base == "" {
		base = filepath.Join(os.Getenv("HOME"), ".cache")
	}
	return filepath.Join(base, "centinela", "update-check.json")
}

// readCache returns the cached tag when a valid, non-expired entry exists. A
// missing, corrupt, empty, or stale cache returns ("", false) without error.
func (u *Updater) readCache() (string, bool) {
	data, err := os.ReadFile(cachePath())
	if err != nil {
		return "", false
	}
	var e cacheEntry
	if err := json.Unmarshal(data, &e); err != nil || e.LatestTag == "" {
		return "", false
	}
	age := u.Now().Sub(time.Unix(e.CheckedAt, 0))
	if age < 0 || age > u.ttl() {
		return "", false
	}
	return e.LatestTag, true
}

// writeCache persists the latest tag with the current timestamp. Best-effort:
// any filesystem error is swallowed so a non-writable cache never blocks.
func (u *Updater) writeCache(tag string) {
	p := cachePath()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return
	}
	data, err := json.Marshal(cacheEntry{LatestTag: tag, CheckedAt: u.Now().Unix()})
	if err != nil {
		return
	}
	_ = os.WriteFile(p, data, 0o644)
}

func (u *Updater) ttl() time.Duration {
	if u.TTL <= 0 {
		return defaultTTL
	}
	return u.TTL
}
