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
| `main` | An untouched mirror of upstream `main`. Only the weekly sync moves it. |
| `Dev` | LuxRail: upstream plus LuxForge changes. The integration branch for all work. |
| `phase-XX-*` | One approved change each, merged into `Dev` by pull request. |
| `sync/upstream-<date>` | A weekly upstream merge, opened as a pull request into `Dev`. |

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
- No GitHub Actions run in this repository: Actions is turned off for it. Builds, tests, syncs and
  releases run on LuxForge devices (Pillars `standards/devices.md`). Upstream's workflow files are
  kept unchanged so merges stay clean; they do not run here.
- Weekly upstream sync, Fridays at 02:00 UTC: a scheduled job on a LuxForge build device fetches upstream, fast-forwards
  `main`, merges it into a `sync/upstream-<date>` branch from `Dev`, runs the tests and opens a pull
  request into `Dev`. Conflicts are left marked in that pull request for a person or agent to
  resolve. Nothing reaches `Dev` without the operator's approval.
