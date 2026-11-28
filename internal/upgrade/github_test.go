// SPDX-License-Identifier: AGPL-3.0-only

package upgrade

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

const testToken = "github_pat_test_only_0123456789"

// apiRelease serves the files of a releaseServer the way the GitHub API
// does for a private repository: release metadata and assets need the
// token, and assets redirect to a separate storage host.
type apiServer struct {
	api, storage *httptest.Server
	status       int // when not 0, every API call answers this
	mu           sync.Mutex
	storageAuth  []string // Authorization headers that reached storage
}

func newAPIServer(t *testing.T, rs *releaseServer, version string) *apiServer {
	t.Helper()
	a := &apiServer{}
	a.storage = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.mu.Lock()
		a.storageAuth = append(a.storageAuth, r.Header.Get("Authorization"))
		a.mu.Unlock()
		b, ok := rs.files["/"+version+"/"+strings.TrimPrefix(r.URL.Path, "/blob/")]
		if !ok {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write(b)
	}))
	t.Cleanup(a.storage.Close)
	a.api = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if a.status != 0 {
			w.WriteHeader(a.status)
			return
		}
		if r.Header.Get("Authorization") != "Bearer "+testToken {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		const repo = "/repos/o/r"
		switch p := strings.TrimPrefix(r.URL.Path, repo); {
		case p == "/releases/latest" || p == "/releases/tags/"+version:
			var assets []map[string]string
			for k := range rs.files {
				if name, ok := strings.CutPrefix(k, "/"+version+"/"); ok {
					assets = append(assets, map[string]string{"name": name, "url": a.api.URL + repo + "/releases/assets/" + name})
				}
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"tag_name": version, "assets": assets})
		case strings.HasPrefix(p, "/releases/assets/"):
			if r.Header.Get("Accept") != "application/octet-stream" {
				http.Error(w, "asset metadata", http.StatusTeapot)
				return
			}
			http.Redirect(w, r, a.storage.URL+"/blob/"+strings.TrimPrefix(p, "/releases/assets/"), http.StatusFound)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(a.api.Close)
	return a
}

func tokenFile(t *testing.T, mode os.FileMode) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "release-token")
	if err := os.WriteFile(p, []byte(testToken+"\n"), mode); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(p, mode); err != nil {
		t.Fatal(err)
	}
	return p
}

// tokenOpts reads releases through a's API with a token; ReleaseBase points
// nowhere, so nothing can come from the public download path.
func tokenOpts(t *testing.T, a *apiServer, o Options) Options {
	o.ReleaseBase = "http://127.0.0.1:1/unused"
	o.LatestURL = "http://127.0.0.1:1/unused/latest.txt"
	o.ReleaseAPI = a.api.URL + "/repos/o/r"
	o.ReleaseTokenFile = tokenFile(t, 0o600)
	return o
}

func TestPrivateReleaseResolveStageAndLatest(t *testing.T) {
	pub, sec := keypair(t)
	rs, _ := newRelease(t, sec, "v0.2.0", map[string]string{"5050d": "#!/bin/sh\n", "5050-agent": "#!/bin/sh\n"})
	a := newAPIServer(t, rs, "v0.2.0")
	o := tokenOpts(t, a, baseOpts(t, rs.URL, pub))
	ctx := context.Background()

	if v, err := Latest(ctx, o); err != nil || v != "v0.2.0" {
		t.Fatalf("latest: %q %v", v, err)
	}
	r, err := Resolve(ctx, o, "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(r.URL, a.api.URL+"/repos/o/r/releases/assets/") {
		t.Fatalf("release URL %s is not an API asset", r.URL)
	}
	if strings.Contains(r.URL, testToken) {
		t.Fatal("the token is in the release URL")
	}
	if _, err := Stage(ctx, o, r, "5050d", "5050-agent"); err != nil {
		t.Fatal(err)
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if len(a.storageAuth) == 0 {
		t.Fatal("nothing was downloaded from storage")
	}
	for _, h := range a.storageAuth {
		if h != "" {
			t.Fatalf("the redirect to storage carried Authorization %q", h)
		}
	}
}

func TestPrivateReleaseRefusals(t *testing.T) {
	pub, sec := keypair(t)
	rs, _ := newRelease(t, sec, "v0.2.0", map[string]string{"5050d": "x"})
	for _, tc := range []struct {
		status int
		code   string
	}{
		{http.StatusUnauthorized, CodeReleaseAuthFailed},
		{http.StatusForbidden, CodeReleaseNotAccessible},
		{http.StatusNotFound, CodeReleaseNotAccessible},
	} {
		t.Run(fmt.Sprint(tc.status), func(t *testing.T) {
			a := newAPIServer(t, rs, "v0.2.0")
			a.status = tc.status
			o := tokenOpts(t, a, baseOpts(t, rs.URL, pub))
			_, err := Resolve(context.Background(), o, "v0.2.0")
			var re *ReleaseError
			if !errors.As(err, &re) || re.Code != tc.code {
				t.Fatalf("error %v, want code %s", err, tc.code)
			}
			msg := err.Error()
			if !strings.Contains(msg, "v0.2.0") || !strings.Contains(msg, fmt.Sprint(tc.status)) || !strings.Contains(msg, o.ReleaseTokenFile) {
				t.Errorf("the error does not name the release, answer and token file: %s", msg)
			}
			if strings.Contains(msg, testToken) {
				t.Error("the error holds the token")
			}
		})
	}
}

func TestPrivateReleaseWrongTokenIsAuthFailure(t *testing.T) {
	pub, sec := keypair(t)
	rs, _ := newRelease(t, sec, "v0.2.0", map[string]string{"5050d": "x"})
	a := newAPIServer(t, rs, "v0.2.0")
	o := tokenOpts(t, a, baseOpts(t, rs.URL, pub))
	if err := os.WriteFile(o.ReleaseTokenFile, []byte("github_pat_wrong"), 0o600); err != nil {
		t.Fatal(err)
	}
	var re *ReleaseError
	if _, err := Resolve(context.Background(), o, "v0.2.0"); !errors.As(err, &re) || re.Code != CodeReleaseAuthFailed {
		t.Fatalf("a wrong token: %v", err)
	}
}

func TestReleaseTokenFileChecks(t *testing.T) {
	if _, err := readToken(tokenFile(t, 0o644)); err == nil || !strings.Contains(err.Error(), "chmod 600") {
		t.Errorf("a world-readable token file: %v", err)
	}
	if _, err := readToken(filepath.Join(t.TempDir(), "none")); err == nil {
		t.Error("a missing token file was read")
	}
	p := tokenFile(t, 0o600)
	if err := os.WriteFile(p, []byte("  \n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := readToken(p); err == nil {
		t.Error("an empty token file was read")
	}
}

func TestReleaseAPIFromBase(t *testing.T) {
	if got := releaseAPIFor(DefaultReleaseBase, DefaultGitHubAPI); got != "https://api.github.com/repos/LuxForge-ltd/luxrail" {
		t.Errorf("default base: %q", got)
	}
	for _, b := range []string{"https://example.com/releases", "http://github.com/o/r/releases/download", "https://github.com/o/r/archive"} {
		if got := releaseAPIFor(b, DefaultGitHubAPI); got != "" {
			t.Errorf("%s gave API %q", b, got)
		}
	}
	// Without a token nothing goes through the API, and "-" turns a token off.
	o := Options{ReleaseFile: "-"}.Defaults()
	if o.usesToken() {
		t.Error("a token is used with none configured")
	}
	if o := (Options{ReleaseFile: "-", ReleaseTokenFile: "-"}).Defaults(); o.ReleaseTokenFile != "" {
		t.Errorf("\"-\" kept token file %q", o.ReleaseTokenFile)
	}
}

func TestPublicReleaseSendsNoToken(t *testing.T) {
	pub, sec := keypair(t)
	rs, _ := newRelease(t, sec, "v0.2.0", map[string]string{"5050d": "x"})
	var seen []string
	inner := rs.Config.Handler
	rs.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = append(seen, r.Header.Get("Authorization"))
		inner.ServeHTTP(w, r)
	})
	o := baseOpts(t, rs.URL, pub)
	o.ReleaseTokenFile = tokenFile(t, 0o600) // set, but ReleaseBase is not GitHub, so no API
	if _, err := Resolve(context.Background(), o, ""); err != nil {
		t.Fatal(err)
	}
	for _, h := range seen {
		if h != "" {
			t.Fatalf("a plain release download carried Authorization %q", h)
		}
	}
}
