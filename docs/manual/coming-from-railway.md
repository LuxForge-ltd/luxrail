---
title: Coming from Railway
description: LuxRail follows Railway names, variables and config files, so most projects move over unchanged; this page lists every difference in v0.1.
---

LuxRail follows Railway names, variables and config files, so most projects move over unchanged; this page lists every difference in v0.1.

## What stays the same

- The canvas, staged changes, environments and the service panel tabs.
- Variable names (`RAILWAY_PUBLIC_DOMAIN`, `RAILWAY_PRIVATE_DOMAIN` and the rest) are provided alongside `FIFTY_*` names, and the reference syntax `${{Service.VAR}}` and `${{shared.VAR}}` is the same.
- `railway.json` and `railway.toml` in your repo are read and applied. See [Config as code](guides/config-as-code.md).
- Builds use Railpack, Railway's builder. `builder: NIXPACKS` builds with Railpack and maps Nixpacks hints.
- Private names: `api.railway.internal` resolves as well as `api.internal`.
- The `railway` command works as an alias of the `luxrail` CLI. See [The railway alias](cli/railway-alias.md).
- Deployment statuses use Railway's labels.

## Differences

| Railway | In LuxRail | What to do instead |
|---|---|---|
| PR environments (incl. Focused and Bot PR environments) | Not yet (v0.2) | Duplicate an environment and point a service at the branch |
| Wait for CI | Not yet | Deploy from CI with `luxrail up -c` or a GitHub Action calling `luxrail redeploy` |
| Deployment approvals | Not yet | Turn off auto deploy and deploy by hand |
| Image Auto Updates with maintenance windows | Not yet | "Check for a newer image" in Settings > Source, or a cron that calls `luxrail redeploy` |
| Serverless (app sleeping) | Not available | Scale to 0 is v0.2; idle apps cost only RAM on your own VM |
| Buckets | Not yet | The MinIO template |
| Functions | Not available | A small service or a cron service |
| Observability widget board | Not yet | The Observability page and Metrics tabs |
| AI Diagnose and fix PR | Rule-based sentence | The diagnosis sentence and the [Troubleshooting guide](guides/troubleshooting.md) |
| Data tab SQL editor and row editing | Read-only tables and Redis keys | `luxrail connect` |
| Metrics for 30 days | Same, 30 days | |
| Wildcard custom domains | Not yet | One custom domain per name, or `--domain` with a wildcard A record for generated names |
| Static egress IPs (paid) | Free: your hosts' IPs | Settings > Networking shows them |
| Volume Live Resize | Not needed: volumes use the host disk | Add a data disk |
| Regions | Your hosts and roles | Add a host in another location |
| `railway` CLI | `luxrail`, plus the `railway` alias | Same commands |
| Login with GitHub, magic link email login | Email and password with 2FA | |
| Slack/Discord formatted webhooks | Signed JSON | A small relay service, or v0.2 |
| `config plan/apply` style config sync | `luxrail config pull` and `railway.json` in the repo | v0.2 Git mode |

## Railway variables that are not provided

Railway's `OTEL_*` tracing variables are not set. Set them yourself to point at your own collector. See [Variables reference](reference/variables.md).

## Moving a project

1. [Move a project from Railway](guides/import-from-railway.md): recreate the services from the same repos and copy their variables. The one-step importer is not in 0.1.0.
2. [Move a database from Railway](guides/move-database-from-railway.md) copies the data.
3. Point your DNS at your host and [add the custom domains](guides/custom-domains.md).
4. Run `luxrail alias install` so `railway` commands talk to your console.

## Next steps

- [Compare: Railway](compare/railway.md)
- [Honest limits](reference/honest-limits.md)
