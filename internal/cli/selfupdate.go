// SPDX-License-Identifier: MIT

package cli

// LuxRail: the CLI keeps itself at the version of the server it talks to. On
// launch (at most once an hour) it asks the server for its version; when the
// server is newer, it downloads the server's CLI build, checks the SHA-256
// the server sends (taken from the release's signed checksums), puts the new
// binary in place of the running one and runs the command again with it.
// LUXRAIL_NO_UPDATE=1 turns this off; `luxrail self-update` runs it now.

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/luxforge-ltd/luxrail/pkg/api"
	"github.com/luxforge-ltd/luxrail/pkg/version"
)

const (
	updateEvery     = time.Hour
	maxCLIArchive   = 256 << 20
	updateCheckFile = "update-check"
)

// noUpdateArgs are first arguments that never trigger an automatic update.
var noUpdateArgs = map[string]bool{
	"self-update": true, "update": true, "version": true, "help": true, "completion": true,
	"login": true, "logout": true, "-h": true, "--help": true, "__complete": true, "__completeNoDesc": true,
}

var releaseRE = regexp.MustCompile(`^v(\d+)\.(\d+)\.(\d+)(-[0-9A-Za-z.-]+)?$`)

// newerVersion reports whether release tag a is newer than b. Anything that
// is not a release tag (a development build) is never newer.
func newerVersion(a, b string) bool {
	pa, oka := parseVersion(a)
	pb, okb := parseVersion(b)
	if !oka || !okb {
		return false
	}
	for i := 0; i < 3; i++ {
		if pa[i] != pb[i] {
			return pa[i] > pb[i]
		}
	}
	// A plain release is newer than a pre-release of the same numbers.
	return pa[3] == 0 && pb[3] == 1
}

func parseVersion(v string) ([4]int, bool) {
	m := releaseRE.FindStringSubmatch(strings.TrimSpace(v))
	if m == nil {
		return [4]int{}, false
	}
	var out [4]int
	for i := 0; i < 3; i++ {
		n, err := strconv.Atoi(m[i+1])
		if err != nil {
			return [4]int{}, false
		}
		out[i] = n
	}
	if m[4] != "" {
		out[3] = 1
	}
	return out, true
}

// updater holds what self-update replaces in tests.
type updater struct {
	exe    func() (string, error) // the running binary
	goos   string
	goarch string
	now    func() time.Time
}

func (a *App) updater() updater {
	if a.upd.exe == nil {
		a.upd.exe = func() (string, error) {
			p, err := os.Executable()
			if err != nil {
				return "", err
			}
			return filepath.EvalSymlinks(p)
		}
	}
	if a.upd.goos == "" {
		a.upd.goos, a.upd.goarch = runtime.GOOS, runtime.GOARCH
	}
	if a.upd.now == nil {
		a.upd.now = time.Now
	}
	return a.upd
}

// autoUpdate runs before a command. It reports whether the CLI was replaced;
// the caller then runs the command again with the new binary.
func (a *App) autoUpdate(args []string) bool {
	if a.env("FIFTY_NO_UPDATE") != "" || !releaseRE.MatchString(version.Version) {
		return false
	}
	if len(args) > 0 && noUpdateArgs[args[0]] {
		return false
	}
	u := a.updater()
	stamp, err := a.updateStamp()
	if err != nil {
		return false
	}
	if st, err := os.Stat(stamp); err == nil && u.now().Sub(st.ModTime()) < updateEvery {
		return false
	}
	cl, _, _, err := a.client()
	if err != nil {
		return false
	}
	touch(stamp, u.now())
	ctx, cancel := ctxTimeout(3 * time.Second)
	var v api.VersionInfo
	err = cl.Do(ctx, "GET", "/version", nil, &v)
	cancel()
	if err != nil || !newerVersion(v.Version, version.Version) {
		return false
	}
	fmt.Fprintf(a.Err, "Updating luxrail %s to %s, the version of %s...\n", version.Version, v.Version, cl.Base)
	ctx, cancel = ctxTimeout(2 * time.Minute)
	defer cancel()
	if err := a.replaceSelf(ctx, cl); err != nil {
		fmt.Fprintln(a.Err, "Could not update luxrail: "+sentence(err.Error())+" Carrying on with "+version.Version+".")
		return false
	}
	fmt.Fprintf(a.Err, "Updated luxrail to %s.\n", v.Version)
	return true
}

// updateStamp is the file whose time records the last automatic check.
func (a *App) updateStamp() (string, error) {
	p, err := configPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Dir(p), updateCheckFile), nil
}

func touch(p string, t time.Time) {
	if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
		return
	}
	if _, err := os.Stat(p); errors.Is(err, fs.ErrNotExist) {
		_ = os.WriteFile(p, nil, 0o600)
	}
	_ = os.Chtimes(p, t, t)
}

// replaceSelf downloads the server's CLI for this platform and puts it in
// place of the running binary.
func (a *App) replaceSelf(ctx context.Context, cl *Client) error {
	u := a.updater()
	exe, err := u.exe()
	if err != nil {
		return fmt.Errorf("cannot find the running luxrail binary: %w", err)
	}
	archive, err := cl.downloadCLI(ctx, u.goos, u.goarch)
	if err != nil {
		return err
	}
	bin, err := binaryFromArchive(archive, u.goos)
	if err != nil {
		return err
	}
	return installBinary(exe, bin, u.goos)
}

// downloadCLI fetches /cli/download/<os>/<arch> and checks its SHA-256.
func (c *Client) downloadCLI(ctx context.Context, goos, goarch string) ([]byte, error) {
	req, err := c.newReq(ctx, http.MethodGet, "/cli/download/"+goos+"/"+goarch, nil, "")
	if err != nil {
		return nil, err
	}
	resp, err := c.send(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	want := strings.ToLower(strings.TrimSpace(resp.Header.Get("X-Checksum-Sha256")))
	if len(want) != 64 {
		return nil, errors.New("the server sent the CLI without its checksum, so it was not used")
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, maxCLIArchive+1))
	if err != nil {
		return nil, fmt.Errorf("the download stopped: %w", err)
	}
	if len(b) > maxCLIArchive {
		return nil, errors.New("the CLI download is larger than any CLI should be")
	}
	sum := sha256.Sum256(b)
	if got := hex.EncodeToString(sum[:]); got != want {
		return nil, fmt.Errorf("the downloaded CLI has SHA-256 %s, but the server says %s, so it was not used", got, want)
	}
	return b, nil
}

// binaryFromArchive returns the luxrail binary inside a release archive: a
// .zip holding luxrail.exe on Windows, a .tar.gz holding luxrail elsewhere.
func binaryFromArchive(archive []byte, goos string) ([]byte, error) {
	if goos == "windows" {
		zr, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
		if err != nil {
			return nil, fmt.Errorf("the CLI download is not a zip file: %w", err)
		}
		for _, f := range zr.File {
			if filepath.Base(f.Name) != "luxrail.exe" || f.FileInfo().IsDir() {
				continue
			}
			rc, err := f.Open()
			if err != nil {
				return nil, err
			}
			defer rc.Close()
			return io.ReadAll(io.LimitReader(rc, maxCLIArchive))
		}
		return nil, errors.New("the CLI download has no luxrail.exe")
	}
	gz, err := gzip.NewReader(bytes.NewReader(archive))
	if err != nil {
		return nil, fmt.Errorf("the CLI download is not a .tar.gz file: %w", err)
	}
	tr := tar.NewReader(gz)
	for {
		h, err := tr.Next()
		if errors.Is(err, io.EOF) {
			return nil, errors.New("the CLI download has no luxrail binary")
		}
		if err != nil {
			return nil, fmt.Errorf("cannot read the CLI download: %w", err)
		}
		if h.Typeflag == tar.TypeReg && filepath.Base(h.Name) == "luxrail" {
			return io.ReadAll(io.LimitReader(tr, maxCLIArchive))
		}
	}
}

// installBinary replaces exe with bin. The new file is written next to exe
// first. Windows cannot overwrite a running .exe but can rename it, so the
// old one becomes <exe>.old and is removed on a later run.
func installBinary(exe string, bin []byte, goos string) error {
	dir := filepath.Dir(exe)
	tmp, err := os.CreateTemp(dir, ".luxrail-update-*")
	if err != nil {
		if errors.Is(err, fs.ErrPermission) {
			return fmt.Errorf("luxrail is in %s, which you cannot write to; run `sudo luxrail self-update`, or reinstall it somewhere you own", dir)
		}
		return err
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }()
	if _, err := tmp.Write(bin); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmpName, 0o755); err != nil {
		return err
	}
	if goos != "windows" {
		return os.Rename(tmpName, exe)
	}
	old := exe + ".old"
	_ = os.Remove(old)
	if err := os.Rename(exe, old); err != nil {
		return fmt.Errorf("cannot move the running luxrail aside: %w", err)
	}
	if err := os.Rename(tmpName, exe); err != nil {
		_ = os.Rename(old, exe)
		return fmt.Errorf("cannot put the new luxrail in place: %w", err)
	}
	return nil
}

// removeOldBinary deletes the <exe>.old a Windows update left behind.
func removeOldBinary() {
	if runtime.GOOS != "windows" {
		return
	}
	if p, err := os.Executable(); err == nil {
		_ = os.Remove(p + ".old")
	}
}

// rerun runs the updated binary with the same arguments and returns its exit
// code. Interrupts reach the child through the shared console, so the parent
// ignores them while it waits.
func rerun(exe string, args []string) int {
	cmd := exec.Command(exe, args...)
	if len(os.Args) > 0 {
		cmd.Args[0] = os.Args[0]
	}
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	cmd.Env = append(os.Environ(), "LUXRAIL_NO_UPDATE=1")
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	defer signal.Stop(ch)
	if err := cmd.Run(); err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			return ee.ExitCode()
		}
		fmt.Fprintln(os.Stderr, "Error: could not run the updated luxrail: "+err.Error())
		return 1
	}
	return 0
}

func (a *App) cmdSelfUpdate() *cobra.Command {
	var check bool
	cmd := &cobra.Command{
		Use:     "self-update",
		Aliases: []string{"update"},
		Short:   "Update this CLI to the version of the server it talks to",
		Long: "Update this CLI to the version of the server it talks to.\n\n" +
			"The CLI also does this by itself, at most once an hour, when it runs a command while logged in.\n" +
			"Set LUXRAIL_NO_UPDATE=1 to turn the automatic update off.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, _, _, err := a.client()
			if err != nil {
				return err
			}
			ctx, cancel := ctxTimeout(2 * time.Minute)
			defer cancel()
			var v api.VersionInfo
			if err := cl.Do(ctx, "GET", "/version", nil, &v); err != nil {
				return err
			}
			dev := !releaseRE.MatchString(version.Version)
			if !newerVersion(v.Version, version.Version) && !(dev && releaseRE.MatchString(v.Version)) {
				a.printf("luxrail %s is up to date with %s (%s).\n", version.Version, cl.Base, v.Version)
				return nil
			}
			if check {
				a.printf("luxrail %s can update to %s from %s. Run `luxrail self-update`.\n", version.Version, v.Version, cl.Base)
				return nil
			}
			if err := a.replaceSelf(ctx, cl); err != nil {
				return err
			}
			if stamp, err := a.updateStamp(); err == nil {
				touch(stamp, a.updater().now())
			}
			a.printf("Updated luxrail %s to %s.\n", version.Version, v.Version)
			return nil
		},
	}
	cmd.Flags().BoolVar(&check, "check", false, "only say whether an update is available")
	return cmd
}
