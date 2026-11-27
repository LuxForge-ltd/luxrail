# LuxRail

LuxRail is LuxForge's self-hosted application platform. Install it on a Linux server you own and it becomes a place to run your apps: deploy services from GitHub, Docker images or a local folder, add databases in one click, wire them together with variables, and read logs, metrics and deploy history from a web console or the `luxrail` CLI.

It gives you the workflow of a hosted platform such as Railway on your own hardware, with no per-seat, per-minute or per-server fees. LuxRail is built and maintained by LuxForge Ltd.

## How it works

- **The control plane** (`5050d`) runs on one host. It holds your projects, environments and variables, serves the console and the API, and plans every deploy.
- **A host agent** (`5050-agent`) runs on every server. It builds images with BuildKit, runs containers, configures Caddy for HTTPS and joins hosts over WireGuard. Agents dial out to the control plane, so no management port is opened.
- **The console and the CLI** talk to the same API. Changes are staged per environment and shown as a diff before they deploy.
- **Releases are signed**, and an upgrade reverts itself if the new version does not come up healthy.

See the [architecture reference](docs/manual/reference/architecture.md) for the full picture.

## What it does

- **Deploys your code**: GitHub repositories (Railpack or Dockerfile) with push-to-deploy, Docker images, or a folder with `luxrail up`.
- **Runs your databases**: Postgres, MySQL, Redis and MongoDB in one click, backups to any S3-compatible bucket, and point-in-time restore for Postgres.
- **Connects services**: variable references such as `${{Postgres.DATABASE_URL}}` and `${{shared.KEY}}`, and private networking between services.
- **Serves the web**: custom domains with automatic HTTPS, TCP proxies, health checks, zero-downtime deploys and rollbacks that restore variables.
- **Runs scheduled jobs**: cron services with a schedule in the console or in `railway.json`.
- **Keeps watch**: logs with filters, 30 days of metrics, monitors and alerts.
- **Grows with you**: add a server with one command, run replicas across hosts, and use every core with one process per vCPU.
- **Moves you off Railway**: `luxrail import railway`, `railway.json` support and a `railway` CLI alias.

Read the [honest limits](docs/manual/reference/honest-limits.md) before you rely on it: one control host, each database on one host, IPv4 only, and Ubuntu 24.04 as the tested system.

## Install

On a fresh Linux server:

```sh
curl -fsSL https://github.com/LuxForge-ltd/luxrail/releases/latest/download/install.sh | sudo sh
```

Piped like this, the installer asks nothing. Run from a terminal, it offers a guided setup. Every answer is also a flag or a `FIFTY_*` environment variable (see [install options](docs/manual/reference/install-options.md)). The script is also in [`install/install.sh`](install/install.sh).

The CLI, on your own machine:

```sh
curl -fsSL https://github.com/LuxForge-ltd/luxrail/releases/latest/download/install.sh | sh -s -- --cli
luxrail login https://console.example.com
luxrail up --new
```

If the console uses its own certificate authority, pass it once with `luxrail login URL --ca-file FILE`. This matters most on Windows, where the CLI only reads the Windows certificate store.

## Documentation

- [Quick start](docs/manual/quick-start.md) and [basics](docs/manual/basics.md)
- [Guides](docs/manual/guides/install.md): deploying, databases, domains, backups, upgrades and more
- [CLI reference](docs/manual/cli/overview.md)
- [Production checklist](docs/manual/production-checklist.md)
- [Coming from Railway](docs/manual/coming-from-railway.md)
- [`LUXRAIL.md`](LUXRAIL.md): branches, versions, releases and LuxForge's own changes

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

To build from source, see [CONTRIBUTING.md](CONTRIBUTING.md).

## License

The server and console are [AGPL-3.0-only](LICENSE). The CLI, `pkg/`, `templates/`, `examples/`, `install/` and `docs/manual/` are MIT. See [LICENSING.md](LICENSING.md) for the details and the copyright notices.

Security issues: see [SECURITY.md](SECURITY.md). LuxRail is an independent project and is not affiliated with Railway.

<sub>Sourced and forked from [openrail](https://github.com/musakkka/openrail) v0.1.9.</sub>
