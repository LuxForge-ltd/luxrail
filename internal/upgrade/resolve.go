// SPDX-License-Identifier: AGPL-3.0-only

package upgrade

import (
	"bufio"
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/luxforge-ltd/luxrail/internal/upgrade/minisign"
)

// Release is one verified release tarball for one architecture.
type Release struct {
	Version string
	URL     string // tarball
	SHA256  string // verified against the signed checksums file
}

var versionRE = regexp.MustCompile(`^v[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?$`)

// ValidVersion reports whether v is a release tag such as v0.1.0 or
// v0.1.0-e2e.2.
func ValidVersion(v string) bool { return versionRE.MatchString(v) }

// ServerTarball is the name of the server tarball of a release
// (5050_0.1.0_linux_amd64.tar.gz for v0.1.0).
func ServerTarball(version, arch string) string {
	return fmt.Sprintf("5050_%s_linux_%s.tar.gz", strings.TrimPrefix(version, "v"), arch)
}

// Resolve finds a release (version "" = latest) and verifies its signed checksum.
func Resolve(ctx context.Context, o Options, version string) (Release, error) {
	o = o.Defaults()
	keys, err := o.keys()
	if err != nil {
		return Release{}, err
	}
	if version == "" || version == "latest" {
		v, err := Latest(ctx, o)
		if err != nil {
			return Release{}, fmt.Errorf("cannot find the latest release: %w", err)
		}
		version = v
	}
	return resolveFile(ctx, o, keys, version, func(v string) string { return ServerTarball(v, o.Arch) })
}

// resolveFile downloads the signed checksums of version, verifies them
// against keys and returns the release file that name gives for it.
func resolveFile(ctx context.Context, o Options, keys []minisign.PublicKey, version string, name func(string) string) (Release, error) {
	if !strings.HasPrefix(version, "v") {
		version = "v" + version
	}
	if !ValidVersion(version) {
		return Release{}, fmt.Errorf("%q is not a release version; use the form v0.1.0", version)
	}
	file := name(version)
	where, err := o.releaseFiles(ctx, version, file)
	if err != nil {
		return Release{}, err
	}
	sums, err := fetch(ctx, o, where["checksums.txt"], 1<<20)
	if err != nil {
		return Release{}, fmt.Errorf("cannot download the checksums of %s: %w", version, err)
	}
	sig, err := fetch(ctx, o, where["checksums.txt.minisig"], 64<<10)
	if err != nil {
		return Release{}, fmt.Errorf("cannot download the signature of %s: %w", version, err)
	}
	if _, _, err := minisign.Verify(keys, sums, sig); err != nil {
		return Release{}, fmt.Errorf("the checksums of %s failed signature verification, so nothing was installed: %w", version, err)
	}
	sum, err := lookupChecksum(sums, file)
	if err != nil {
		return Release{}, fmt.Errorf("release %s: %w", version, err)
	}
	return Release{Version: version, URL: where[file], SHA256: sum}, nil
}

// releaseFiles returns where the checksums, their signature and file of
// version are downloaded from: under ReleaseBase, or the release's API
// asset addresses when a token is in use.
func (o Options) releaseFiles(ctx context.Context, version, file string) (map[string]string, error) {
	names := []string{"checksums.txt", "checksums.txt.minisig", file}
	out := map[string]string{}
	if !o.usesToken() {
		for _, n := range names {
			out[n] = o.ReleaseBase + "/" + version + "/" + n
		}
		return out, nil
	}
	r, err := o.apiRelease(ctx, version)
	if err != nil {
		return nil, err
	}
	for _, n := range names[:2] {
		if out[n], err = r.asset(n); err != nil {
			return nil, err
		}
	}
	// A missing build for this architecture is reported from the checksums.
	out[file], _ = r.asset(file)
	return out, nil
}

// lookupChecksum finds name in a checksums.txt ("<sha256>  <file>" lines).
func lookupChecksum(sums []byte, name string) (string, error) {
	sc := bufio.NewScanner(bytes.NewReader(sums))
	for sc.Scan() {
		f := strings.Fields(sc.Text())
		if len(f) != 2 {
			continue
		}
		if strings.TrimPrefix(f[1], "*") != name {
			continue
		}
		if b, err := hex.DecodeString(f[0]); err != nil || len(b) != 32 {
			return "", fmt.Errorf("the checksum of %s is not a SHA-256 digest", name)
		}
		return strings.ToLower(f[0]), nil
	}
	return "", fmt.Errorf("there is no %s in its checksums, so this release has no build for this architecture", name)
}

// fetch GETs url and returns at most limit bytes.
func fetch(ctx context.Context, o Options, url string, limit int64) ([]byte, error) {
	resp, err := o.openAsset(ctx, url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(b)) > limit {
		return nil, errors.New(url + " is larger than expected")
	}
	return b, nil
}
