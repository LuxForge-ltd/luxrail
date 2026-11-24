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

## Phases

| Phase | Branch | Status | Version | Pull request | Commit approval |
| --- | --- | --- | --- | --- | --- |
| 01 | `phase-01-fork-governance-setup` | Open | 0.0.1 | see pull requests | Plan approved by Ben, 2026-11-24 18:47 UTC ("lets do it. Call it LuxRail"); commit approved by Ben, 2026-11-24 18:51 UTC ("go luxrail 01"); merge not authorised |

## Handoff Notes

- Phase 01 adds LuxForge governance (`AGENTS.md`, `CLAUDE.md`, `.github/copilot-instructions.md`, `.claude/`, `.gitattributes` from the Pillars sync script), this ledger, `VERSION` 0.0.1 and `LUXRAIL.md`. No upstream file is changed.
- Upstream remote: `https://github.com/musakkka/openrail.git` (add it as `upstream` in a fresh clone). Base: v0.1.9, `d6fad37`.
- Upstream's `.github/workflows` are kept unchanged for clean merges; Actions is turned off for this repository, so they never run.
- Joyeuse has no Go toolchain yet; the build and release phase must install Go (and confirm Node) before `make build`, `make test` or `make web`.
