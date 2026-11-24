# Phase Ledger

Handoff record for agent work in this repository, as required by LuxForge [AI Agent Governance](https://github.com/LuxForge-ltd/pillars/blob/main/standards/ai-agent-governance.md#phase-ledger). Read this before starting work; update it as part of each approved commit.

- Repository: `LuxForge-ltd/luxrail`
- Integration branch: `Dev`
- Current version: see `VERSION`

## Approved Plans

### LuxRail: a maintained LuxForge fork of openrail

- 2026-11-24 18:47 UTC, Ben (Luxhome project thread "Publish OpenRail at openrail.luxforge.dev"): "lets do it. Call it LuxRail", approving the plan sent at 18:46 UTC in the same thread.
- The plan: fork `musakkka/openrail` as public `LuxForge-ltd/luxrail` (AGPL); `main` mirrors upstream, `Dev` carries LuxForge changes; a weekly upstream sync job on a LuxForge build device (no GitHub Actions) that opens a pull request into `Dev` and leaves conflicts marked for resolution; builds and minisign-signed releases on Joyeuse so installs can upgrade through `--release-base`; then one phase per change: empty lists returned as `null` (console pages stuck loading), a LAN/IP console that survives Caddy rebuilds, no forced HTTPS redirect for tunnelled hosts, switching an existing install to private mode, the LuxForge theme, and other fixes Ben names.
- Done when the plan was approved: fork created 2026-11-24 ~18:49 UTC from upstream `main` at `d6fad37` (openrail v0.1.9); `Dev` created at the same commit; GitHub Actions turned off for the repository.
- 2026-11-24 18:51 UTC, Ben (same thread): "go luxrail 01, sync friday 02:00, upstream yes". The weekly upstream sync runs Fridays at 02:00 UTC, and fixes that are useful to everyone are also offered upstream as pull requests.
- 2026-11-24 18:55 UTC, Ben (same thread), on how the sync works: "Create a new branch, ensure merges work - push to Dev. Leave a task for me in the morning to approve push to main"; run it "on Savathun, via an OpenRail deployment - we can keep an eye on it then"; conflicts as a draft pull request: "ok". So a clean, tested upstream merge is pushed to `Dev` by the job itself (a standing authorisation for that job only), `main` moves only when Ben approves the "Release Dev to main" pull request, and the job runs as a cron service on the OpenRail install on VM 105 (Savathun) rather than on a build device. The phase numbers after 02 move up by one: builds and signed releases become phase 03.

## Phases

| Phase | Branch | Status | Version | Pull request | Commit approval |
| --- | --- | --- | --- | --- | --- |
| 01 | `phase-01-fork-governance-setup` | Merged | 0.0.1 | [#1](https://github.com/LuxForge-ltd/luxrail/pull/1) | Plan approved by Ben, 2026-11-24 18:47 UTC ("lets do it. Call it LuxRail"); commit approved by Ben, 2026-11-24 18:51 UTC ("go luxrail 01"); merge authorised by Ben, 2026-11-24 19:07 UTC ("merge both") |
| 02 | `phase-02-upstream-sync-job` | Merged | 0.0.2 | see pull requests | Plan approved by Ben, 2026-11-24 18:47 UTC; design settled 18:55 UTC (see Approved Plans); commit and merge approved by Ben, 2026-11-24 19:07 UTC ("go luxrail 02, merge both") |

## Handoff Notes

- Phase 01 adds LuxForge governance (`AGENTS.md`, `CLAUDE.md`, `.github/copilot-instructions.md`, `.claude/`, `.gitattributes` from the Pillars sync script), this ledger, `VERSION` 0.0.1 and `LUXRAIL.md`. No upstream file is changed.
- Upstream remote: `https://github.com/musakkka/openrail.git` (add it as `upstream` in a fresh clone). Base: v0.1.9, `d6fad37`.
- Upstream's `.github/workflows` are kept unchanged for clean merges; Actions is turned off for this repository, so they never run.
- Phase 02 (stacked on phase 01's branch, because PR #1 is not merged yet) adds `luxforge/sync/`: `sync-upstream.sh`, its `Dockerfile` (Go 1.26, Node 24, git, gh; user `luxrail`), `railway.json` for the OpenRail cron service (`0 2 * * 5`, Dockerfile builder, no restarts), a runbook and `test/test-sync-upstream.sh` (local bare repositories and a fake `gh`). Checked on joyeuse in Docker: 29 of 29 tests pass, shellcheck clean, the image builds and has every tool, and openrail's own `pkg/configfile` parses `railway.json`. Deploying the service needs a fine-grained GitHub token (`luxrail/github-sync-token` in the vault) and Ben's go.
- Follow-up for Pillars: `standards/devices.md` should list the OpenRail/LuxRail install on VM 105 as where LuxRail's weekly sync runs.
- Joyeuse has no Go toolchain yet; the build and release phase must install Go (and confirm Node) before `make build`, `make test` or `make web`.
