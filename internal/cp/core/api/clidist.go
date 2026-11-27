// SPDX-License-Identifier: AGPL-3.0-only

package api

// LuxRail: the server hands out the luxrail CLI of its own version, for the
// console's Install CLI window and for the CLI's self-update. Archives come
// from the release, checked against its signed checksums, and are kept under
// <DataDir>/cli/<version>/.

import (
	"context"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/luxforge-ltd/luxrail/internal/cp/core/fail"
	"github.com/luxforge-ltd/luxrail/internal/upgrade"
	"github.com/luxforge-ltd/luxrail/pkg/api"
)

// CLIDist hands out the CLI of Version. A nil CLIDist, or one whose Version
// is not a release (a development build), hands out nothing.
type CLIDist struct {
	Opts    upgrade.Options
	Version string

	mu    sync.Mutex
	files map[string]cliFile // "os/arch"
}

type cliFile struct{ path, sha string }

func (d *CLIDist) available() bool { return d != nil && upgrade.ValidVersion(d.Version) }

func (d *CLIDist) info() api.CLIInfo {
	out := api.CLIInfo{Downloads: []api.CLIDownload{}}
	if d == nil {
		return out
	}
	out.Version = d.Version
	out.Available = d.available()
	if !out.Available {
		return out
	}
	for _, p := range upgrade.CLIPlatforms {
		out.Downloads = append(out.Downloads, api.CLIDownload{
			OS: p.OS, Arch: p.Arch, File: upgrade.CLIArchive(d.Version, p.OS, p.Arch),
			Path: "/api/v1/cli/download/" + p.OS + "/" + p.Arch,
		})
	}
	return out
}

// get returns the verified archive for goos/goarch, downloading it once.
func (d *CLIDist) get(ctx context.Context, goos, goarch string) (cliFile, error) {
	if !d.available() {
		return cliFile{}, fail.NotFound("This server is a development build, so it has no CLI to hand out. Build the CLI from source or install it from a release.")
	}
	if !upgrade.HasCLI(goos, goarch) {
		return cliFile{}, fail.NotFound("There is no luxrail CLI for %s/%s.", goos, goarch)
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	key := goos + "/" + goarch
	if f, ok := d.files[key]; ok {
		if _, err := os.Stat(f.path); err == nil {
			return f, nil
		}
	}
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Minute)
	defer cancel()
	r, err := upgrade.ResolveCLI(ctx, d.Opts, d.Version, goos, goarch)
	if err != nil {
		return cliFile{}, fail.Unavailable("Cannot get the %s CLI for %s/%s from the release: %s", d.Version, goos, goarch, err.Error())
	}
	path, err := upgrade.FetchCLI(ctx, d.Opts, r)
	if err != nil {
		return cliFile{}, fail.Unavailable("Cannot download the %s CLI for %s/%s: %s", d.Version, goos, goarch, err.Error())
	}
	if d.files == nil {
		d.files = map[string]cliFile{}
	}
	f := cliFile{path: path, sha: r.SHA256}
	d.files[key] = f
	return f, nil
}

func (s *Server) cliInfo(_ *Ctx) (any, error) { return s.CLI.info(), nil }

func (s *Server) cliDownload(x *Ctx) (any, error) {
	f, err := s.CLI.get(x.R.Context(), x.p("os"), x.p("arch"))
	if err != nil {
		return nil, err
	}
	fh, err := os.Open(f.path)
	if err != nil {
		return nil, fail.Internal(err)
	}
	defer fh.Close()
	st, err := fh.Stat()
	if err != nil {
		return nil, fail.Internal(err)
	}
	name := upgrade.CLIArchive(s.CLI.Version, x.p("os"), x.p("arch"))
	h := x.W.Header()
	h.Set("Content-Type", "application/octet-stream")
	h.Set("Content-Disposition", `attachment; filename="`+name+`"`)
	h.Set("X-Checksum-Sha256", f.sha)
	h.Set("Cache-Control", "private, max-age=300")
	http.ServeContent(x.W, x.R, name, st.ModTime(), fh)
	return handled{}, nil
}
