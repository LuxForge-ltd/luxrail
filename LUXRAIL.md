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

### Moving VM 105 from `caddy-lan` to `consoleLan`

VM 105 serves `https://10.10.180.5:8443` from a separate Caddy (`openrail-lan.service`,
`/etc/caddy-lan/Caddyfile`, its own internal CA). Once VM 105 runs a release with `consoleLan`,
and with the operator's go and a backup first:

1. Stop and disable `openrail-lan.service`, which holds port 8443.
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
