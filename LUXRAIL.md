# LuxRail

LuxRail is LuxForge's maintained fork of [openrail](https://github.com/musakkka/openrail), the
self-hosted PaaS. It tracks upstream closely and carries a small set of LuxForge changes: fixes we
need before upstream ships them, LuxForge integration (Cloudflare tunnels, LAN access) and the
LuxForge look.

Upstream's own documentation (`README.md`, `docs/`) applies unless this file says otherwise. Those
files are left as upstream wrote them, so weekly syncs merge cleanly.

## Licence

openrail is licensed under the GNU Affero General Public License v3.0, with some directories under
MIT (see `LICENSING.md`). LuxRail keeps those licences. Because LuxRail is served over a network,
this repository stays public so everyone using a LuxRail install can get its source.

## Branches

| Branch | Holds |
| --- | --- |
| `main` | Released LuxRail. It moves only when the operator approves the "Release Dev to main" pull request. |
| `Dev` | LuxRail: upstream plus LuxForge changes. The integration branch for all work. |
| `phase-XX-*` | One approved change each, merged into `Dev` by pull request. |
| `sync/upstream-<date>` | A weekly upstream merge. When it is clean and the tests pass, `Dev` is fast-forwarded to it. |

The other branches copied from upstream when the fork was made are upstream's work in progress and
are not used here.

## Versions

LuxRail's own version is in `VERSION` (SemVer, starting at 0.0.1; each phase adds one to PATCH).
The upstream release it is based on is recorded in `PHASES.md` with each sync. Upstream started
LuxRail at openrail v0.1.9 (`d6fad37`).

## Working rules

- Agent and contributor rules: `AGENTS.md` (LuxForge governance, integration branch `Dev`).
- Keep LuxForge changes small and, where possible, in files of their own, so upstream merges stay
  cheap. Fixes that are useful to everyone may be offered upstream as pull requests.
- No GitHub Actions run in this repository: Actions is turned off for it. Builds, tests and
  releases run on LuxForge devices (Pillars `standards/devices.md`); the weekly sync runs on LuxForge's own LuxRail install. Upstream's workflow files are
  kept unchanged so merges stay clean; they do not run here.
- Weekly upstream sync, Fridays at 02:00 UTC, as a cron service on LuxForge's OpenRail/LuxRail
  install: it merges upstream `main` into a `sync/upstream-<date>` branch from `Dev` and runs the
  tests. A clean, passing merge fast-forwards `Dev` and leaves a "Release Dev to main" pull request
  for the operator to approve; conflicts or failing tests leave a pull request into `Dev` instead.
  See `luxforge/sync/README.md`.

## How LuxRail behaves differently from openrail

- **HTTPS-forwarded requests are served on port 80.** openrail answers plain HTTP for an `acme`
  or `internal` domain with a 308 to `https://`. A Cloudflare tunnel, or another proxy that
  terminates TLS, that forwards to Caddy over plain HTTP would loop on that redirect. LuxRail
  serves a request on port 80 when it carries `X-Forwarded-Proto: https` (which Cloudflare sets
  for visitors that used HTTPS) and redirects the rest as before. A tunnel can therefore point at
  `http://<host>:80` with the app's domain as well as straight at the service.
  (`internal/agent/netw/caddytunnel.go`)
- **A LAN console address.** `"consoleLan": "<IPv4>:<port>"` in `/etc/5050/config.json` (for
  example `10.10.180.5:8443`) serves the console and API on that address of the control host over
  HTTPS, with a certificate for the IP from openrail's Caddy internal CA. It is part of the Caddy
  config openrail renders, so it survives every rebuild, unlike a route added through Caddy's admin
  API. 5050d refuses to start with a value that is not an IPv4 address and port, or that uses
  80, 443, 5050 or 5051. Restart 5050d after changing it. Clients that pin a CA pin the internal
  CA's root, `/var/lib/5050/caddy/data/pki/authorities/local/root.crt`.
  (`internal/cp/core/config/consolelan.go`, `internal/cp/orch/engine/consolelan.go`,
  `internal/agent/netw/caddyconsolelan.go`)

### Moving VM 105 from `caddy-lan` to `consoleLan`

VM 105 serves `https://10.10.180.5:8443` from a separate Caddy (`openrail-lan.service`,
`/etc/caddy-lan/Caddyfile`, its own internal CA). Once VM 105 runs a release with `consoleLan`,
and with the operator's go and a backup first:

1. Stop and disable `openrail-lan.service`, which holds port 8443.
2. Add `"consoleLan": "10.10.180.5:8443"` to `/etc/5050/config.json` and restart 5050d.
3. Clients that pinned `openrail-lan-8443-ca.crt` (the sync and publish scripts on joyeuse, the
   LuxHome Tower) pin the openrail Caddy CA instead (`openrail-trial-ca.crt`; the Tower's bundle
   already holds both).
- **Switching an existing install to private mode.** openrail's installer refuses to make a normal
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
