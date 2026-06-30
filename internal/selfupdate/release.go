package selfupdate

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type asset struct {
	Name string `json:"name"`
	URL  string `json:"browser_download_url"`
}

// Release is the parsed subset of the GitHub releases/latest payload.
type Release struct {
	Tag    string  `json:"tag_name"`
	Assets []asset `json:"assets"`
}

// assetURL returns the download URL for the named asset, if present.
func (r *Release) assetURL(name string) (string, bool) {
	for _, a := range r.Assets {
		if a.Name == name {
			return a.URL, true
		}
	}
	return "", false
}

// resolveLatest fetches and parses the latest release for the repo. A transport
// failure is a network error; any non-200 status or undecodable body is an API
// error. Both are typed and leave the binary untouched.
func (u *Updater) resolveLatest() (*Release, error) {
	url := u.APIBase + "/repos/" + repoSlug + "/releases/latest"
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, newErr(KindNetwork, "build release request", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := u.HTTP.Do(req)
	if err != nil {
		return nil, newErr(KindNetwork, "contact GitHub", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, newErr(KindAPI, fmt.Sprintf("GitHub API returned status %d", resp.StatusCode), nil)
	}
	var rel Release
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, newErr(KindAPI, "decode release payload", err)
	}
	return &rel, nil
}
