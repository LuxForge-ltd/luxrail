// SPDX-License-Identifier: AGPL-3.0-only

package upgrade

// Private releases. When the LuxRail repository is private, release files
// cannot be fetched from github.com/.../releases/download without a token,
// so with ReleaseTokenFile set the release is read through the GitHub REST
// API instead: the release by tag (or the latest release), then each asset
// by its API address with "Accept: application/octet-stream". The token is
// read from its file on every use, sent only to the API (never to the
// storage host GitHub redirects the asset to) and never logged or put in a
// URL.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// DefaultReleaseTokenFile holds a GitHub token that can read the releases of
// a private repository (fine-grained, Contents read-only). It is used when
// it exists and nothing else names a token file.
const DefaultReleaseTokenFile = "/etc/5050/release-token"

// DefaultGitHubAPI is the GitHub REST API.
const DefaultGitHubAPI = "https://api.github.com"

// Error codes of ReleaseError.
const (
	CodeReleaseAuthFailed    = "release_auth_failed"    // GitHub refused the token (401)
	CodeReleaseNotAccessible = "release_not_accessible" // the token cannot see the release (403, 404)
)

// ReleaseError is a release GitHub would not hand over with the token. It
// names the release, the answer and the token file, never the token.
type ReleaseError struct {
	Code      string
	Release   string // tag, or "latest"
	Status    string // "401 Unauthorized"
	TokenFile string
}

func (e *ReleaseError) Error() string {
	if e.Code == CodeReleaseAuthFailed {
		return fmt.Sprintf("GitHub refused the release token in %s for release %s (%s). The token is wrong, expired or revoked; put a valid one in that file", e.TokenFile, e.Release, e.Status)
	}
	return fmt.Sprintf("the release token in %s cannot read release %s (%s). Check the release exists and the token has Contents read access to the repository", e.TokenFile, e.Release, e.Status)
}

// usesToken reports whether releases are read through the API with a token.
func (o Options) usesToken() bool { return o.ReleaseTokenFile != "" && o.ReleaseAPI != "" }

// releaseAPIFor returns the API address of the repository whose releases
// base is a github.com download base, or "".
func releaseAPIFor(base, api string) string {
	u, err := url.Parse(base)
	if err != nil || u.Scheme != "https" || u.Host != "github.com" {
		return ""
	}
	p := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(p) != 4 || p[2] != "releases" || p[3] != "download" || p[0] == "" || p[1] == "" {
		return ""
	}
	return strings.TrimRight(api, "/") + "/repos/" + p[0] + "/" + p[1]
}

// readToken reads the token file. The file must be readable by its owner
// only, so a token is never left readable by every account on the host.
func readToken(path string) (string, error) {
	st, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("cannot read the release token file %s: %w", path, err)
	}
	if st.Mode().Perm()&0o077 != 0 {
		return "", fmt.Errorf("the release token file %s can be read by other accounts; run chmod 600 %s", path, path)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("cannot read the release token file %s: %w", path, err)
	}
	t := strings.TrimSpace(string(b))
	if t == "" || strings.ContainsAny(t, " \t\r\n") {
		return "", fmt.Errorf("the release token file %s does not hold one token", path)
	}
	return t, nil
}

// apiClient is o.HTTP with redirects that never carry the token: GitHub
// sends asset downloads to a signed storage address.
func (o Options) apiClient() *http.Client {
	c := *o.HTTP
	c.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) >= 10 {
			return errors.New("too many redirects")
		}
		req.Header.Del("Authorization")
		return nil
	}
	return &c
}

// apiGet GETs an API address with the token and returns the response; a
// refusal becomes a ReleaseError for release.
func (o Options) apiGet(ctx context.Context, addr, accept, release string) (*http.Response, error) {
	token, err := readToken(o.ReleaseTokenFile)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, addr, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", accept)
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	resp, err := o.apiClient().Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusOK {
		return resp, nil
	}
	_ = resp.Body.Close()
	e := &ReleaseError{Code: CodeReleaseNotAccessible, Release: release, Status: resp.Status, TokenFile: o.ReleaseTokenFile}
	switch resp.StatusCode {
	case http.StatusUnauthorized:
		e.Code = CodeReleaseAuthFailed
	case http.StatusForbidden, http.StatusNotFound:
	default:
		return nil, fmt.Errorf("GitHub answered %s for release %s", resp.Status, release)
	}
	return nil, e
}

// apiRelease is the part of a GitHub release this package reads.
type apiRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"assets"`
}

// apiRelease reads release tag ("latest" for the latest release).
func (o Options) apiRelease(ctx context.Context, tag string) (apiRelease, error) {
	addr := o.ReleaseAPI + "/releases/tags/" + url.PathEscape(tag)
	if tag == "latest" {
		addr = o.ReleaseAPI + "/releases/latest"
	}
	resp, err := o.apiGet(ctx, addr, "application/vnd.github+json", tag)
	if err != nil {
		return apiRelease{}, err
	}
	defer resp.Body.Close()
	var r apiRelease
	if err := json.NewDecoder(io.LimitReader(resp.Body, 4<<20)).Decode(&r); err != nil {
		return apiRelease{}, fmt.Errorf("cannot read release %s from GitHub: %w", tag, err)
	}
	return r, nil
}

// asset returns the API address of the asset called name.
func (r apiRelease) asset(name string) (string, error) {
	for _, a := range r.Assets {
		if a.Name == name {
			return a.URL, nil
		}
	}
	return "", fmt.Errorf("release %s has no %s", r.TagName, name)
}

// isAPIAsset reports whether addr is an asset address of the release API,
// the only addresses the token is sent to.
func (o Options) isAPIAsset(addr string) bool {
	return o.usesToken() && strings.HasPrefix(addr, o.ReleaseAPI+"/releases/assets/")
}

// openAsset opens addr for download: an API asset with the token, anything
// else as a plain GET.
func (o Options) openAsset(ctx context.Context, addr string) (*http.Response, error) {
	if o.isAPIAsset(addr) {
		return o.apiGet(ctx, addr, "application/octet-stream", addr[strings.LastIndex(addr, "/")+1:])
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, addr, nil)
	if err != nil {
		return nil, err
	}
	resp, err := o.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		_ = resp.Body.Close()
		return nil, fmt.Errorf("%s answered %s", addr, resp.Status)
	}
	return resp, nil
}

// Latest returns the latest stable release tag: latest.txt, or the latest
// release through the API when a token is in use.
func Latest(ctx context.Context, o Options) (string, error) {
	o = o.Defaults()
	if o.usesToken() {
		r, err := o.apiRelease(ctx, "latest")
		if err != nil {
			return "", err
		}
		return r.TagName, nil
	}
	b, err := fetch(ctx, o, o.LatestURL, 256)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(b)), nil
}
