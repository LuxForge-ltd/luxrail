// SPDX-License-Identifier: AGPL-3.0-only

package upgrade

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/luxforge-ltd/luxrail/internal/upgrade/minisign"
)

// cliRelease serves version with one Windows CLI archive, signed by key.
func cliRelease(t *testing.T, key minisign.PrivateKey, version string, archive []byte) *releaseServer {
	t.Helper()
	rs, _ := newRelease(t, key, version, map[string]string{"5050d": "x"})
	name := CLIArchive(version, "windows", "amd64")
	sum := sha256.Sum256(archive)
	sums := string(rs.files["/"+version+"/checksums.txt"]) + fmt.Sprintf("%s  %s\n", hex.EncodeToString(sum[:]), name)
	sig, err := minisign.Sign(key, []byte(sums), "test")
	if err != nil {
		t.Fatal(err)
	}
	rs.files["/"+version+"/checksums.txt"] = []byte(sums)
	rs.files["/"+version+"/checksums.txt.minisig"] = []byte(sig)
	rs.files["/"+version+"/"+name] = archive
	return rs
}

func TestCLIArchiveNames(t *testing.T) {
	if got := CLIArchive("v1.0.10", "windows", "amd64"); got != "luxrail-cli_1.0.10_windows_amd64.zip" {
		t.Errorf("windows: %s", got)
	}
	if got := CLIArchive("v1.0.10", "darwin", "arm64"); got != "luxrail-cli_1.0.10_darwin_arm64.tar.gz" {
		t.Errorf("darwin: %s", got)
	}
	if HasCLI("windows", "arm64") || !HasCLI("linux", "arm64") {
		t.Error("HasCLI does not match the release builds")
	}
}

func TestResolveAndFetchCLI(t *testing.T) {
	pub, sec := keypair(t)
	archive := []byte("zip bytes of the windows CLI")
	rs := cliRelease(t, sec, "v1.0.11", archive)
	o := baseOpts(t, rs.URL, pub)
	ctx := context.Background()

	r, err := ResolveCLI(ctx, o, "v1.0.11", "windows", "amd64")
	if err != nil {
		t.Fatal(err)
	}
	path, err := FetchCLI(ctx, o, r)
	if err != nil {
		t.Fatal(err)
	}
	if b, _ := os.ReadFile(path); string(b) != string(archive) {
		t.Fatalf("fetched %q", b)
	}
	// A cached file with the right checksum is reused without a download.
	delete(rs.files, "/v1.0.11/"+CLIArchive("v1.0.11", "windows", "amd64"))
	if again, err := FetchCLI(ctx, o, r); err != nil || again != path {
		t.Fatalf("cached fetch: %s %v", again, err)
	}

	if _, err := ResolveCLI(ctx, o, "v1.0.11", "windows", "arm64"); err == nil {
		t.Error("a platform releases do not build was resolved")
	}
	if _, err := ResolveCLI(ctx, o, "v1.0.11", "linux", "amd64"); err == nil || !strings.Contains(err.Error(), "luxrail-cli_1.0.11_linux_amd64.tar.gz") {
		t.Errorf("a missing archive: %v", err)
	}
}

func TestFetchCLIRefusesWrongChecksum(t *testing.T) {
	pub, sec := keypair(t)
	rs := cliRelease(t, sec, "v1.0.11", []byte("good"))
	o := baseOpts(t, rs.URL, pub)
	ctx := context.Background()
	r, err := ResolveCLI(ctx, o, "v1.0.11", "windows", "amd64")
	if err != nil {
		t.Fatal(err)
	}
	rs.files["/v1.0.11/"+CLIArchive("v1.0.11", "windows", "amd64")] = []byte("tampered")
	if _, err := FetchCLI(ctx, o, r); err == nil || !strings.Contains(err.Error(), "SHA-256") {
		t.Fatalf("a tampered archive was kept: %v", err)
	}
}

func TestResolveCLIRefusesUnsignedChecksums(t *testing.T) {
	_, sec := keypair(t)
	other, _ := keypair(t)
	rs := cliRelease(t, sec, "v1.0.11", []byte("zip"))
	o := baseOpts(t, rs.URL, other)
	if _, err := ResolveCLI(context.Background(), o, "v1.0.11", "windows", "amd64"); err == nil || !strings.Contains(err.Error(), "signature") {
		t.Fatalf("checksums signed by an unknown key were trusted: %v", err)
	}
}
