// SPDX-License-Identifier: MIT

package cli

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/luxforge-ltd/luxrail/pkg/api"
	"github.com/luxforge-ltd/luxrail/pkg/version"
)

func TestNewerVersion(t *testing.T) {
	for _, tc := range []struct {
		a, b string
		want bool
	}{
		{"v1.0.11", "v1.0.10", true},
		{"v1.0.10", "v1.0.11", false},
		{"v1.0.10", "v1.0.10", false},
		{"v2.0.0", "v1.9.9", true},
		{"v1.0.10", "v1.0.10-rc.1", true},
		{"v1.0.10-rc.1", "v1.0.10", false},
		{"v1.0.11", "dev", false},
		{"dev", "v1.0.10", false},
		{"v1.0.4-2-g114c911", "v1.0.3", true},
	} {
		if got := newerVersion(tc.a, tc.b); got != tc.want {
			t.Errorf("newerVersion(%q, %q) = %v, want %v", tc.a, tc.b, got, tc.want)
		}
	}
}

func zipWith(t *testing.T, name, body string) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	_, _ = zw.Create("LICENSE")
	w, err := zw.Create(name)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = w.Write([]byte(body))
	_ = zw.Close()
	return buf.Bytes()
}

func tgzWith(t *testing.T, name, body string) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	_ = tw.WriteHeader(&tar.Header{Name: "LICENSE", Mode: 0o644, Typeflag: tar.TypeReg})
	_ = tw.WriteHeader(&tar.Header{Name: name, Mode: 0o755, Size: int64(len(body)), Typeflag: tar.TypeReg})
	_, _ = tw.Write([]byte(body))
	_ = tw.Close()
	_ = gz.Close()
	return buf.Bytes()
}

func TestBinaryFromArchive(t *testing.T) {
	if b, err := binaryFromArchive(zipWith(t, "luxrail.exe", "WIN"), "windows"); err != nil || string(b) != "WIN" {
		t.Errorf("zip: %q %v", b, err)
	}
	if b, err := binaryFromArchive(tgzWith(t, "luxrail", "NIX"), "linux"); err != nil || string(b) != "NIX" {
		t.Errorf("tar.gz: %q %v", b, err)
	}
	if _, err := binaryFromArchive(zipWith(t, "other.exe", "x"), "windows"); err == nil {
		t.Error("a zip without luxrail.exe was accepted")
	}
	if _, err := binaryFromArchive([]byte("not an archive"), "darwin"); err == nil {
		t.Error("garbage was accepted")
	}
}

// updateServer serves /version as serverVersion and the CLI archive for
// linux/amd64, with sum as its checksum header ("" = the real one).
func updateServer(t *testing.T, serverVersion string, archive []byte, sum string) (*httptest.Server, *int) {
	t.Helper()
	downloads := 0
	if sum == "" {
		s := sha256.Sum256(archive)
		sum = hex.EncodeToString(s[:])
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/version", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(api.VersionInfo{Version: serverVersion})
	})
	mux.HandleFunc("GET /api/v1/cli/download/linux/amd64", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer "+tok {
			http.Error(w, "no", http.StatusUnauthorized)
			return
		}
		downloads++
		w.Header().Set("X-Checksum-Sha256", sum)
		_, _ = w.Write(archive)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv, &downloads
}

// updateApp is an App logged in to srv whose "running binary" is a file in a
// temporary directory.
func updateApp(t *testing.T, srv string, now time.Time) (*App, string, *bytes.Buffer) {
	t.Helper()
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "cfg", "config.json")
	t.Setenv("LUXRAIL_CONFIG", cfgPath)
	t.Setenv("LUXRAIL_NO_UPDATE", "")
	t.Setenv("FIFTY_NO_UPDATE", "")
	c := &Config{path: cfgPath, Instances: map[string]string{srv: tok}, Default: srv}
	if err := c.Save(); err != nil {
		t.Fatal(err)
	}
	exe := filepath.Join(dir, "bin", "luxrail")
	_ = os.MkdirAll(filepath.Dir(exe), 0o755)
	if err := os.WriteFile(exe, []byte("OLD"), 0o755); err != nil {
		t.Fatal(err)
	}
	var errBuf bytes.Buffer
	a := &App{Out: &bytes.Buffer{}, Err: &errBuf, Cwd: dir,
		upd: updater{exe: func() (string, error) { return exe, nil }, goos: "linux", goarch: "amd64", now: func() time.Time { return now }}}
	return a, exe, &errBuf
}

func withVersion(t *testing.T, v string) {
	t.Helper()
	old := version.Version
	version.Version = v
	t.Cleanup(func() { version.Version = old })
}

func TestAutoUpdateReplacesTheBinaryOnceAnHour(t *testing.T) {
	withVersion(t, "v1.0.10")
	srv, downloads := updateServer(t, "v1.0.11", tgzWith(t, "luxrail", "NEW"), "")
	now := time.Date(2026, 11, 27, 20, 0, 0, 0, time.UTC)
	a, exe, errBuf := updateApp(t, srv.URL, now)

	if !a.autoUpdate([]string{"status"}) {
		t.Fatalf("no update: %s", errBuf)
	}
	if b, _ := os.ReadFile(exe); string(b) != "NEW" {
		t.Fatalf("binary is %q", b)
	}
	if !strings.Contains(errBuf.String(), "Updated luxrail to v1.0.11") {
		t.Errorf("message: %s", errBuf)
	}
	// Within the hour nothing is checked again; after it, the check runs again.
	a.upd.now = func() time.Time { return now.Add(30 * time.Minute) }
	if a.autoUpdate([]string{"status"}) || *downloads != 1 {
		t.Errorf("checked again within the hour (%d downloads)", *downloads)
	}
	a.upd.now = func() time.Time { return now.Add(61 * time.Minute) }
	if !a.autoUpdate([]string{"status"}) || *downloads != 2 {
		t.Errorf("no check after the hour (%d downloads)", *downloads)
	}
}

func TestAutoUpdateSkips(t *testing.T) {
	now := time.Date(2026, 11, 27, 20, 0, 0, 0, time.UTC)
	t.Run("server not newer", func(t *testing.T) {
		withVersion(t, "v1.0.11")
		srv, downloads := updateServer(t, "v1.0.11", tgzWith(t, "luxrail", "NEW"), "")
		a, exe, _ := updateApp(t, srv.URL, now)
		if a.autoUpdate([]string{"status"}) || *downloads != 0 {
			t.Fatal("updated to the same version")
		}
		if b, _ := os.ReadFile(exe); string(b) != "OLD" {
			t.Fatal("binary changed")
		}
	})
	t.Run("development build", func(t *testing.T) {
		withVersion(t, "dev")
		srv, downloads := updateServer(t, "v1.0.11", tgzWith(t, "luxrail", "NEW"), "")
		a, _, _ := updateApp(t, srv.URL, now)
		if a.autoUpdate([]string{"status"}) || *downloads != 0 {
			t.Fatal("a development build updated itself")
		}
	})
	t.Run("turned off", func(t *testing.T) {
		withVersion(t, "v1.0.10")
		srv, downloads := updateServer(t, "v1.0.11", tgzWith(t, "luxrail", "NEW"), "")
		a, _, _ := updateApp(t, srv.URL, now)
		t.Setenv("LUXRAIL_NO_UPDATE", "1")
		if a.autoUpdate([]string{"status"}) || *downloads != 0 {
			t.Fatal("LUXRAIL_NO_UPDATE was ignored")
		}
	})
	t.Run("version and login commands", func(t *testing.T) {
		withVersion(t, "v1.0.10")
		srv, downloads := updateServer(t, "v1.0.11", tgzWith(t, "luxrail", "NEW"), "")
		a, _, _ := updateApp(t, srv.URL, now)
		for _, c := range []string{"version", "login", "self-update"} {
			if a.autoUpdate([]string{c}) {
				t.Errorf("%s updated", c)
			}
		}
		if *downloads != 0 {
			t.Fatal("downloaded")
		}
	})
	t.Run("wrong checksum", func(t *testing.T) {
		withVersion(t, "v1.0.10")
		srv, _ := updateServer(t, "v1.0.11", tgzWith(t, "luxrail", "NEW"), strings.Repeat("0", 64))
		a, exe, errBuf := updateApp(t, srv.URL, now)
		if a.autoUpdate([]string{"status"}) {
			t.Fatal("a download with the wrong checksum was used")
		}
		if b, _ := os.ReadFile(exe); string(b) != "OLD" {
			t.Fatal("binary changed")
		}
		if !strings.Contains(errBuf.String(), "Could not update luxrail") {
			t.Errorf("message: %s", errBuf)
		}
	})
}

func TestInstallBinaryOnWindowsKeepsTheOldOneAside(t *testing.T) {
	dir := t.TempDir()
	exe := filepath.Join(dir, "luxrail.exe")
	_ = os.WriteFile(exe, []byte("OLD"), 0o755)
	_ = os.WriteFile(exe+".old", []byte("OLDER"), 0o755)
	if err := installBinary(exe, []byte("NEW"), "windows"); err != nil {
		t.Fatal(err)
	}
	if b, _ := os.ReadFile(exe); string(b) != "NEW" {
		t.Errorf("exe is %q", b)
	}
	if b, _ := os.ReadFile(exe + ".old"); string(b) != "OLD" {
		t.Errorf("old is %q", b)
	}
}

func TestSelfUpdateCommand(t *testing.T) {
	withVersion(t, "v1.0.10")
	srv, _ := updateServer(t, "v1.0.11", tgzWith(t, "luxrail", "NEW"), "")
	a, exe, _ := updateApp(t, srv.URL, time.Now())
	out := &bytes.Buffer{}
	a.Out = out
	if code := a.Run("luxrail", []string{"self-update", "--check"}); code != 0 {
		t.Fatalf("check exit %d", code)
	}
	if !strings.Contains(out.String(), "can update to v1.0.11") {
		t.Errorf("check: %s", out)
	}
	if b, _ := os.ReadFile(exe); string(b) != "OLD" {
		t.Fatal("--check changed the binary")
	}
	out.Reset()
	if code := a.Run("luxrail", []string{"self-update"}); code != 0 {
		t.Fatalf("exit %d", code)
	}
	if b, _ := os.ReadFile(exe); string(b) != "NEW" || !strings.Contains(out.String(), "Updated luxrail v1.0.10 to v1.0.11") {
		t.Fatalf("binary %q, output %s", b, out)
	}
}
