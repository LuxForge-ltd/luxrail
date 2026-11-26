# LuxRail

Railway's developer experience on servers you own.

One command turns any Linux VM into a canvas of services with git deploys, databases, logs and rollbacks. Open source. No per-seat, per-minute or per-server fees.

```sh
curl -fsSL https://github.com/LuxForge-ltd/luxrail/releases/latest/download/install.sh | sudo sh
```

Or with wget: `wget -qO- https://github.com/LuxForge-ltd/luxrail/releases/latest/download/install.sh | sudo sh`. Piped, the installer asks nothing; run from a terminal, it offers a guided setup. Every answer is also a flag or a `FIFTY_*` environment variable. The same script is at [`install/install.sh`](install/install.sh).

Docs: the [user manual](docs/manual/quick-start.md) and [`LUXRAIL.md`](LUXRAIL.md).

## Features

- **Canvas, environments and staged changes**: the Railway workflow, with a diff before every deploy.
- **Deploy anything**: GitHub repos (Railpack or Dockerfile), Docker images, or a directory with `luxrail up`.
- **Databases in one click**: Postgres, MySQL, Redis and MongoDB, with Railway's variable names.
- **Variables with references**: `${{Postgres.DATABASE_URL}}` and `${{shared.KEY}}`, the same syntax as Railway.
- **Backups to any S3-compatible bucket**: point-in-time restore for Postgres, restore to a copy, promote.
- **Zero-downtime deploys** for services without a volume, health checks, and rollbacks that restore variables.
- **More hosts with one command**: agents dial out, so no inbound management port; replicas across hosts; private networking over WireGuard.
- **Use all cores**: run one process per vCPU, with Node cluster mode built in.
- **Custom domains with automatic HTTPS**, TCP proxies, logs with Railway's filter syntax, metrics for 30 days, monitors and alerts.
- **Move from Railway**: `luxrail import railway`, `railway.json` honoured, and a `railway` CLI alias.
- **Self-reverting upgrades** and control plane state streamed to your bucket.

Read the [honest limits](docs/manual/reference/honest-limits.md) before you rely on it: one control host, databases on one host each, IPv4 only, Ubuntu 24.04 tested.

## CLI

```sh
curl -fsSL https://github.com/LuxForge-ltd/luxrail/releases/latest/download/install.sh | sh -s -- --cli
luxrail login https://console.example.com
luxrail up --new
```

See the [CLI reference](docs/manual/cli/overview.md).

## Repository

| Path | What |
|---|---|
| `cmd/5050d`, `internal/cp` | Control plane and console API |
| `cmd/5050-agent`, `internal/agent` | Host agent: runtime, builds, networking, ingress |
| `cmd/luxrail`, `internal/cli` | The CLI |
| `web/` | The console |
| `pkg/` | API, agent protocol and manifest types |
| `templates/` | Database and app templates |
| `install/` | The installer and systemd units |
| `docs/` | The user manual (`docs/manual/`) and the design spec |
| `luxforge/` | Release tooling |
| `e2e/`, `examples/` | End-to-end tests and sample apps |

Build from source: see [CONTRIBUTING.md](CONTRIBUTING.md).

## License

The server and console are [AGPL-3.0-only](LICENSE). The CLI, `pkg/`, `templates/`, `examples/`, `install/` and `docs/manual/` are MIT. See [LICENSING.md](LICENSING.md).

LuxRail is maintained by LuxForge Ltd. It is a modified version of openrail by The openrail authors, forked from openrail v0.1.9 on 2026-11-24 and changed by LuxForge Ltd since then; their copyright notices are kept in the licence files.

Security issues: see [SECURITY.md](SECURITY.md). LuxRail is an independent project and is not affiliated with Railway.
