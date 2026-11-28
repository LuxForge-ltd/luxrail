# LuxRail

LuxRail is LuxForge's self-hosted PaaS. It is a modified version of openrail by The openrail
authors, forked at openrail v0.1.9 (`d6fad37`) on 2026-11-24, and is developed on its own: it does
not follow, merge from or contribute back to the original project.

The documentation in `README.md` and `docs/` applies unless this file says otherwise.

## Licence

LuxRail is licensed under the GNU Affero General Public License v3.0, with some directories under
MIT (see `LICENSING.md`), as the original was, and keeps The openrail authors' copyright notices. Because LuxRail is served over a network,
this repository stays public so everyone using a LuxRail install can get its source.

## Branches

| Branch | Holds |
| --- | --- |
| `main` | Released LuxRail. It moves only when the operator approves the "Release Dev to main" pull request. |
| `Dev` | The integration branch for all work. |
| `phase-XX-*` | One approved change each, merged into `Dev` by pull request. |

Other branches and the `v0.1.0` to `v0.1.9` tags were copied when the fork was made and are not
used.

## Versions

LuxRail's version is in `VERSION` (SemVer; each phase adds one to PATCH). Releases are tagged with
the same number, `v<VERSION>`, starting at `v1.0.0` (see `luxforge/release/README.md`).

## Working rules

- Agent and contributor rules: `AGENTS.md` (LuxForge governance, integration branch `Dev`).
- No GitHub Actions run in this repository, and it has no workflow files. Builds, tests and
  releases run on LuxForge devices (Pillars `standards/devices.md`).
- `main` moves when the operator merges a "Release Dev to main" pull request.

## LuxForge additions

- **HTTPS-forwarded requests are served on port 80.** LuxRail used to answer plain HTTP for an `acme`
  or `internal` domain with a 308 to `https://`. A Cloudflare tunnel, or another proxy that
  terminates TLS, that forwards to Caddy over plain HTTP would loop on that redirect. LuxRail
  serves a request on port 80 when it carries `X-Forwarded-Proto: https` (which Cloudflare sets
  for visitors that used HTTPS) and redirects the rest. A tunnel can therefore point at
  `http://<host>:80` with the app's domain as well as straight at the service.
  (`internal/agent/netw/caddytunnel.go`)
- **A LAN console address.** `"consoleLan": "<IPv4>:<port>"` in `/etc/5050/config.json` (for
  example `10.10.180.5:8443`) serves the console and API on that address of the control host over
  HTTPS, with a certificate for the IP from LuxRail's Caddy internal CA. It is part of the Caddy
  config LuxRail renders, so it survives every rebuild, unlike a route added through Caddy's admin
  API. 5050d refuses to start with a value that is not an IPv4 address and port, or that uses
  80, 443, 5050 or 5051. Restart 5050d after changing it. Clients that pin a CA pin the internal
  CA's root, `/var/lib/5050/caddy/data/pki/authorities/local/root.crt`.
  (`internal/cp/core/config/consolelan.go`, `internal/cp/orch/engine/consolelan.go`,
  `internal/agent/netw/caddyconsolelan.go`)

- **Private console names and IPs use the internal CA.** A wildcard DNS name such as
  `console.10-10-180-5.sslip.io` (sslip.io or nip.io) that can only point at a private, loopback
  or link-local IPv4 address, and a bare IP, get a certificate from LuxRail's Caddy internal CA
  instead of ACME, which could never issue one. When the console's own name is such a name, the
  console is also served on that bare IP (`https://10.10.180.5`) with an internal certificate.
  Both are part of the Caddy config LuxRail renders, so a reboot or agent restart no longer drops
  them, as it did the routes once added by hand with the `caddy_ip_site.py` and
  `caddy_console_split.py` helpers. (`internal/cp/orch/routes/privateip.go`,
  `internal/cp/orch/engine/consoleip.go`)
- **A link address.** `"linkUrl": "https://luxrail.example.com"` in `/etc/5050/config.json` makes
  the links LuxRail hands out (CLI login approval, invites, password resets, email verification,
  notification emails and upload log links) start with that address, for example a tunnel name
  that works from anywhere. Unlike `consoleUrl`, it leaves the console's name in Caddy alone.
  5050d refuses to start with a value that is not `https://` and a host. Restart 5050d after
  changing it. (`internal/cp/core/config/linkurl.go`)
- **The CLI from the console, and a CLI that updates itself.** `GET /api/v1/cli` and
  `/api/v1/cli/download/{os}/{arch}` hand out the CLI of the server's own version: 5050d downloads
  the archive from the release once, checks it against the signed checksums, keeps it under
  `<dataDir>/cli/<version>/` and serves it to any signed-in user. The user menu's **Install CLI**
  window uses them. The CLI checks the server's version at most once an hour when it runs a
  command, and when the server is newer it replaces itself with the server's build (checking the
  SHA-256 the server sends) and runs the command again. `luxrail self-update` does it now and
  `LUXRAIL_NO_UPDATE=1` turns it off. (`internal/upgrade/cli.go`, `internal/cp/core/api/clidist.go`,
  `internal/cli/selfupdate.go`, `web/src/features/cli/`)

- **Tailwind CSS 4.** The console builds with Tailwind 4 through `@tailwindcss/vite` (no PostCSS or
  `tailwind.config.js`); the theme lives in `@theme` in `web/src/index.css`, and LuxRail's base and
  custom rules sit in cascade layers so utilities still win. The console now needs Chrome or Edge
  111, Safari 16.4 or Firefox 128 or later. Two Tailwind 4 rules to keep in mind: a plain
  `hidden` no longer beats `inline`, `inline-block` or `inline-flex` in the same class list (use
  `max-sm:hidden` and the like), and `rounded-sm` is now 4px (Tailwind 3's 2px is `rounded-xs`).
  `web/e2e/specs/visual/screens.e2e.ts` and `scripts/screens-diff.py` capture every main screen
  in both themes at three widths and compare two builds image by image.

### Moving VM 105 from `caddy-lan` to `consoleLan`

VM 105 serves `https://10.10.180.5:8443` from a separate Caddy (`luxrail-lan.service`,
`/etc/caddy-lan/Caddyfile`, its own internal CA). Once VM 105 runs a release with `consoleLan`,
and with the operator's go and a backup first:

1. Stop and disable `luxrail-lan.service`, which holds port 8443.
2. Add `"consoleLan": "10.10.180.5:8443"` to `/etc/5050/config.json` and restart 5050d.
3. Clients that pinned `openrail-lan-8443-ca.crt` (the sync and publish scripts on joyeuse, the
   LuxHome Tower) pin LuxRail's Caddy CA instead (`openrail-trial-ca.crt`; the Tower's bundle
   already holds both).
- **Switching an existing install to private mode.** The installer used to refuse to make a normal
  install private. LuxRail adds `5050d make-private` (sets `"private": true` in `config.json` and,
  with `--base-domain` and `--console-domain`, replaces sslip.io or nip.io names in `config.json`
  and `state.db`; it checks everything before writing, and `--undo` reverses the flag) and the
  installer flag `--to-private`, which runs it, restarts 5050d, starts the tunnel, proves the
  console answers through it, and only then closes the ports. If any check fails, the run puts the
  install back as it was. `--keep-console-from CIDR` keeps the LAN console (`consoleLan`)
  reachable from a range once the private firewall is on, as `--keep-ssh-from` does for SSH.
  (`cmd/5050d/private.go`, `install/install.sh`)

  ```
  sudo FIFTY_TUNNEL_TOKEN=... sh install.sh --to-private --release-base <LuxRail release base> \
    --domain apps.example.com --console-domain console.example.com \
    --keep-ssh-from 10.0.0.0/16 --keep-console-from 10.0.0.0/16
  ```

  Before running it on VM 105: take a vzdump and `5050d backup-state`; the console then goes
  through `5050-tunnel.service`, which replaces the hand-installed host `cloudflared` (use the same
  tunnel token, then stop the old service); and the private firewall closes everything else,
  including the old `caddy-lan` on 8443, so move to `consoleLan` first.
- **The LuxForge look.** The console shows the LuxForge mark (`web/src/kit/LuxForgeMark.tsx`,
  also `web/public/favicon.svg`), the name LuxRail in its header and tab titles
  (`web/src/kit/brand.ts`), and the LuxForge Framework's plum brand scale and page colours
  (`web/src/kit/luxforge.css`, loaded after `theme.css`, which is unchanged). The install name shown in the console is a setting (Settings > General > Name).

## Services LuxForge runs on LuxRail

- **A trial Tower** (`luxforge/tower/`): LuxForge's Tower, installed from the Foundry feed at
  `TOWER_VERSION`, as a service in `luxrail-ops` with its web frontend on a LuxRail domain, so its
  behaviour can be watched on a real install. `luxforge/tower/README.md` says how to deploy it.
