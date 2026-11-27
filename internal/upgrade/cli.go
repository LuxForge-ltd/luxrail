// SPDX-License-Identifier: AGPL-3.0-only

package upgrade

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// CLIPlatform is one operating system and architecture a release builds the
// luxrail CLI for.
type CLIPlatform struct{ OS, Arch string }

// CLIPlatforms are the CLI builds of every release (.goreleaser.yaml, id cli).
var CLIPlatforms = []CLIPlatform{
	{"windows", "amd64"}, {"darwin", "arm64"}, {"darwin", "amd64"}, {"linux", "amd64"}, {"linux", "arm64"},
}

// HasCLI reports whether releases build the CLI for goos/goarch.
func HasCLI(goos, goarch string) bool {
	for _, p := range CLIPlatforms {
		if p.OS == goos && p.Arch == goarch {
			return true
		}
	}
	return false
}

// CLIArchive is the name of the CLI archive of a release
// (luxrail-cli_1.0.10_windows_amd64.zip for v1.0.10 on Windows).
func CLIArchive(version, goos, goarch string) string {
	ext := ".tar.gz"
	if goos == "windows" {
		ext = ".zip"
	}
	return fmt.Sprintf("luxrail-cli_%s_%s_%s%s", strings.TrimPrefix(version, "v"), goos, goarch, ext)
}

// ResolveCLI finds the CLI archive of version for goos/goarch and verifies
// its signed checksum, as Resolve does for the server tarball.
func ResolveCLI(ctx context.Context, o Options, version, goos, goarch string) (Release, error) {
	o = o.Defaults()
	if !HasCLI(goos, goarch) {
		return Release{}, fmt.Errorf("releases have no CLI for %s/%s", goos, goarch)
	}
	keys, err := o.keys()
	if err != nil {
		return Release{}, err
	}
	return resolveFile(ctx, o, keys, version, func(v string) string { return CLIArchive(v, goos, goarch) })
}

// FetchCLI downloads the CLI archive r into <DataDir>/cli/<version>/ and
// returns its path. A file already there with the signed checksum is reused;
// anything else is downloaded again and checked before it is kept.
func FetchCLI(ctx context.Context, o Options, r Release) (string, error) {
	o = o.Defaults()
	if !ValidVersion(r.Version) {
		return "", fmt.Errorf("%q is not a release version", r.Version)
	}
	dir := filepath.Join(o.DataDir, "cli", r.Version)
	path := filepath.Join(dir, filepath.Base(r.URL))
	if sum, err := fileSHA256(path); err == nil && strings.EqualFold(sum, r.SHA256) {
		return path, nil
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	tmp, err := os.CreateTemp(dir, ".download-*")
	if err != nil {
		return "", err
	}
	defer func() {
		_ = tmp.Close()
		_ = os.Remove(tmp.Name())
	}()
	if err := download(ctx, o.HTTP, r.URL, tmp, r.SHA256); err != nil {
		return "", err
	}
	if err := tmp.Close(); err != nil {
		return "", err
	}
	if err := os.Chmod(tmp.Name(), 0o644); err != nil {
		return "", err
	}
	if err := os.Rename(tmp.Name(), path); err != nil {
		return "", err
	}
	return path, nil
}

func fileSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
