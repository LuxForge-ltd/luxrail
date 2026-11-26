// SPDX-License-Identifier: AGPL-3.0-only

// LuxRail: 5050d make-private turns an existing install into a private one.
//
// Purpose: LuxRail can only be installed private from scratch; its
// installer refuses an existing normal host. A private install needs two
// things on the control plane's side: "private": true in config.json, and
// base and console domains that are real names (not sslip.io or nip.io),
// both in config.json and in state.db, where the console edits them. This
// command makes both changes, checking everything before it writes
// anything. The tunnel and the firewall are the installer's part
// (install.sh --to-private runs this command first).
// Scope: the control host's config.json and state.db. It does not restart
// 5050d, start a tunnel or close any port.
// Operational role: run by the installer, or by hand followed by a restart
// of 5050d. --undo sets "private" back to false.
package main

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/luxforge-ltd/luxrail/internal/cp/core"
	"github.com/luxforge-ltd/luxrail/internal/cp/core/config"
	"github.com/luxforge-ltd/luxrail/pkg/api"
)

// makePrivateFlags are the command line of make-private.
type makePrivateFlags struct {
	config, baseDomain, consoleDomain string
	undo                              bool
}

// parseMakePrivate reads the flags of make-private.
//
// Parameters: args is the command line after "make-private". Returns the
// flags or a usage error.
func parseMakePrivate(args []string) (makePrivateFlags, error) {
	var m makePrivateFlags
	fs := newFlags("make-private")
	cfg := configFlag(fs)
	fs.StringVar(&m.baseDomain, "base-domain", "", "base domain for apps, a domain in Cloudflare with a wildcard route to the tunnel (needed when the current one is an sslip.io or nip.io name)")
	fs.StringVar(&m.consoleDomain, "console-domain", "", "console domain, routed through the tunnel (needed when the current one is an sslip.io or nip.io name)")
	fs.BoolVar(&m.undo, "undo", false, `set "private" back to false`)
	if err := fs.Parse(args); err != nil {
		return m, err
	}
	m.config = *cfg
	m.baseDomain = strings.ToLower(strings.TrimSpace(m.baseDomain))
	m.consoleDomain = strings.ToLower(strings.TrimSpace(m.consoleDomain))
	return m, nil
}

// cmdMakePrivate is "5050d make-private".
//
// Parameters: ctx bounds the state.db work; args are the flags; stdout gets
// what was changed and what to do next.
func cmdMakePrivate(ctx context.Context, args []string, stdout io.Writer) error {
	m, err := parseMakePrivate(args)
	if err != nil {
		return err
	}
	f, err := config.Load(m.config)
	if err != nil {
		return fmt.Errorf("cannot read %s: %w", m.config, err)
	}
	if m.undo {
		return undoPrivate(m.config, f, stdout)
	}
	if f.Private {
		_, _ = fmt.Fprintln(stdout, "This install is already private; nothing was changed.")
		return nil
	}
	return makePrivate(ctx, m, f, stdout)
}

// makePrivate checks the private config and the domains in state.db, then
// writes state.db (domains) and config.json (private, domains).
//
// Parameters: m are the flags; f is the loaded, non-private config.json.
func makePrivate(ctx context.Context, m makePrivateFlags, f *config.File, stdout io.Writer) error {
	next := *f
	next.Private = true
	if m.baseDomain != "" {
		next.BaseDomain = m.baseDomain
	}
	if m.consoleDomain != "" {
		next.ConsoleDomain = m.consoleDomain
	}
	if err := next.Validate(); err != nil {
		return fmt.Errorf("nothing was changed: %w; pass --base-domain and --console-domain with names in Cloudflare", err)
	}
	p, err := openWith(ctx, m.config, &next, nil, nil)
	if err != nil {
		return fmt.Errorf("nothing was changed: %w", err)
	}
	defer p.Close()
	req, err := privateDomainChanges(ctx, p.c, m)
	if err != nil {
		return err
	}
	if req.BaseDomain != nil || req.ConsoleDomain != nil {
		if _, err := p.c.UpdateInstall(ctx, core.Actor{Email: "5050d make-private"}, req); err != nil {
			return fmt.Errorf("nothing was changed: %w", err)
		}
	}
	if err := config.Save(m.config, &next); err != nil {
		return fmt.Errorf("the domains were updated in state.db, but %s could not be written: %w", m.config, err)
	}
	_, _ = fmt.Fprintf(stdout, "%s now says private, with base domain %s and console domain %s.\n", m.config, next.BaseDomain, next.ConsoleDomain)
	_, _ = fmt.Fprintln(stdout, "Restart 5050d (systemctl restart 5050d), then start the tunnel and close the ports with install.sh --to-private, or undo with 5050d make-private --undo.")
	return nil
}

// privateDomainChanges decides which domains in state.db must change:
// a flag always wins, and a domain that is an sslip.io or nip.io name with
// no flag to replace it is refused, because a private install cannot
// serve it.
//
// Parameters: c is the control plane opened with the private config; m
// are the flags. Returns the update to apply (possibly empty).
func privateDomainChanges(ctx context.Context, c *core.Core, m makePrivateFlags) (api.UpdateInstallRequest, error) {
	var req api.UpdateInstallRequest
	in, err := c.Install(ctx)
	if err != nil {
		return req, fmt.Errorf("nothing was changed: cannot read the install settings: %w", err)
	}
	console := consoleDomainOf(in.ConsoleURL)
	for _, d := range []struct{ flag, cur, name string }{
		{m.baseDomain, in.BaseDomain, "base"},
		{m.consoleDomain, console, "console"},
	} {
		if d.flag == "" && config.IPNameDomain(d.cur) {
			return req, fmt.Errorf("nothing was changed: the %s domain %s is an sslip.io or nip.io name, which a private install cannot serve; pass --%s-domain with a name in Cloudflare", d.name, d.cur, d.name)
		}
	}
	if m.baseDomain != "" && m.baseDomain != in.BaseDomain {
		req.BaseDomain = &m.baseDomain
	}
	if m.consoleDomain != "" && m.consoleDomain != console {
		req.ConsoleDomain = &m.consoleDomain
	}
	return req, nil
}

// consoleDomainOf returns the host name of a console URL such as
// "https://console.example.com".
func consoleDomainOf(u string) string {
	h := strings.TrimPrefix(strings.TrimPrefix(u, "https://"), "http://")
	h, _, _ = strings.Cut(h, "/")
	h, _, _ = strings.Cut(h, ":")
	return strings.ToLower(h)
}

// undoPrivate sets "private" back to false in config.json. Domains are left
// as they are: real names work on a normal install too.
//
// Parameters: path and f are config.json and its contents.
func undoPrivate(path string, f *config.File, stdout io.Writer) error {
	if !f.Private {
		_, _ = fmt.Fprintln(stdout, "This install is not private; nothing was changed.")
		return nil
	}
	f.Private = false
	if err := config.Save(path, f); err != nil {
		return err
	}
	_, _ = fmt.Fprintf(stdout, "%s no longer says private. Restart 5050d (systemctl restart 5050d).\n", path)
	return nil
}
