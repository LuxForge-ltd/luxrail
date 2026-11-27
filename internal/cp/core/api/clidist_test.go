// SPDX-License-Identifier: AGPL-3.0-only

// LuxRail: the server hands out the CLI of its own version.
package api

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/luxforge-ltd/luxrail/internal/upgrade"
	"github.com/luxforge-ltd/luxrail/internal/upgrade/minisign"
	"github.com/luxforge-ltd/luxrail/pkg/api"
)

// cliReleaseServer serves v1.0.11 with a signed Windows CLI archive and
// counts the archive downloads.
func cliReleaseServer(t *testing.T, archive []byte) (*httptest.Server, minisign.PublicKey, *int) {
	t.Helper()
	pub, sec, err := minisign.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	name := upgrade.CLIArchive("v1.0.11", "windows", "amd64")
	sum := sha256.Sum256(archive)
	sums := fmt.Sprintf("%s  %s\n", hex.EncodeToString(sum[:]), name)
	sig, err := minisign.Sign(sec, []byte(sums), "test")
	if err != nil {
		t.Fatal(err)
	}
	gets := 0
	files := map[string][]byte{
		"/v1.0.11/checksums.txt": []byte(sums), "/v1.0.11/checksums.txt.minisig": []byte(sig), "/v1.0.11/" + name: archive,
	}
	rs := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, ok := files[r.URL.Path]
		if !ok {
			http.NotFound(w, r)
			return
		}
		if r.URL.Path == "/v1.0.11/"+name {
			gets++
		}
		_, _ = w.Write(b)
	}))
	t.Cleanup(rs.Close)
	return rs, pub, &gets
}

func TestCLIDownloads(t *testing.T) {
	archive := []byte("the windows CLI zip")
	rs, pub, gets := cliReleaseServer(t, archive)
	e := newEnv(t)
	s := New(e.c, nil, nil)
	s.CLI = &CLIDist{Version: "v1.0.11", Opts: upgrade.Options{ReleaseBase: rs.URL, DataDir: t.TempDir(),
		ReleaseFile: "-", PublicKeys: []string{pub.String()}}}
	srv := httptest.NewServer(s)
	t.Cleanup(srv.Close)
	e.srv = srv

	anon := e.client()
	anon.do("GET", "/api/v1/cli", nil, nil, 401)
	owner := e.client()
	tok, err := e.c.CreateSetupToken(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	owner.do("POST", "/api/v1/setup", api.SetupRequest{Token: tok, Email: "owner@example.com", Name: "Owner", Password: "correct horse battery"}, nil, 200)
	viewer, _ := join(t, e, owner, api.RoleViewer, "viewer@example.com")

	var info api.CLIInfo
	viewer.do("GET", "/api/v1/cli", nil, &info, 200)
	if info.Version != "v1.0.11" || !info.Available || len(info.Downloads) != len(upgrade.CLIPlatforms) {
		t.Fatalf("info: %+v", info)
	}
	if d := info.Downloads[0]; d.OS != "windows" || d.File != "luxrail-cli_1.0.11_windows_amd64.zip" || d.Path != "/api/v1/cli/download/windows/amd64" {
		t.Fatalf("first download: %+v", d)
	}

	for i := 0; i < 2; i++ {
		req, _ := http.NewRequest("GET", srv.URL+"/api/v1/cli/download/windows/amd64", nil)
		resp, err := viewer.h.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		b, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		sum := sha256.Sum256(archive)
		if resp.StatusCode != 200 || string(b) != string(archive) || resp.Header.Get("X-Checksum-Sha256") != hex.EncodeToString(sum[:]) {
			t.Fatalf("download %d: %d %q %q", i, resp.StatusCode, b, resp.Header.Get("X-Checksum-Sha256"))
		}
		if cd := resp.Header.Get("Content-Disposition"); cd != `attachment; filename="luxrail-cli_1.0.11_windows_amd64.zip"` {
			t.Fatalf("content disposition: %s", cd)
		}
	}
	if *gets != 1 {
		t.Errorf("the archive was fetched from the release %d times, want once", *gets)
	}
	viewer.do("GET", "/api/v1/cli/download/windows/arm64", nil, nil, 404)
	viewer.do("GET", "/api/v1/cli/download/linux/amd64", nil, nil, 503)
}

func TestCLIDownloadsOnADevelopmentBuild(t *testing.T) {
	e := newEnv(t)
	owner := e.client()
	tok, err := e.c.CreateSetupToken(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	owner.do("POST", "/api/v1/setup", api.SetupRequest{Token: tok, Email: "owner@example.com", Name: "Owner", Password: "correct horse battery"}, nil, 200)
	var info api.CLIInfo
	owner.do("GET", "/api/v1/cli", nil, &info, 200)
	if info.Available || len(info.Downloads) != 0 {
		t.Fatalf("a server without a CLI offered one: %+v", info)
	}
	owner.do("GET", "/api/v1/cli/download/windows/amd64", nil, nil, 404)
}
