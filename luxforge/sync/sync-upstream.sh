#!/usr/bin/env bash
# sync-upstream.sh -- merge upstream openrail into LuxRail's Dev branch, once a week.
#
# Runs as the `upstream-sync` cron service on LuxForge's OpenRail/LuxRail install, Fridays at 02:00 UTC
# (the schedule is set on the service; see README.md next to this file), or by hand with the same
# variables.
#
#   1. Clone LuxRail and fetch upstream `main`.
#   2. Nothing new upstream: log "up to date" and stop.
#   3. Branch `sync/upstream-<date>` from Dev and merge upstream `main` into it.
#   4. Conflicts: commit the merge with its conflict markers, push the branch and open a draft pull
#      request into Dev titled "needs resolution" that lists the files. Dev is not touched.
#   5. Clean merge: run the tests. If they fail, push the branch and open a pull request into Dev
#      with the failure; Dev is not touched.
#   6. Tests pass: push the branch, then fast-forward Dev to it (never forced), and make sure a
#      pull request from Dev into main is open and assigned, so the operator can approve the
#      release to main in the morning.
#
# Variables:
#   GH_TOKEN           token with contents and pull-request write access to LUXRAIL_REPO (secret)
#   LUXRAIL_REPO       default LuxForge-ltd/luxrail
#   UPSTREAM_REPO      default musakkka/openrail
#   UPSTREAM_BRANCH    default main
#   SYNC_ASSIGNEE      GitHub login assigned to the pull requests (default 30-something-programmer)
#   SYNC_TEST_CMD      test command run in the merged tree (default: see DEFAULT_TEST_CMD)
#   SYNC_WORKDIR       scratch directory (default: a new temporary directory)
#   SYNC_GIT_URL_BASE  base URL for clones (default https://github.com; tests use a local path)
#   SYNC_DATE          date in branch names, YYYY-MM-DD (default: today, UTC)
#
# Exit codes: 0 up to date or synced into Dev; 10 conflicts (draft PR opened); 11 tests failed
# (PR opened); 12 Dev moved during the run (branch pushed, Dev left alone); other = error.
# The token is passed to git through an askpass helper and is never printed or written to disk
# outside the scratch directory, which is removed on exit.
# Single-quoted strings below hold Markdown backticks for pull request text, not shell expansions.
# shellcheck disable=SC2016
set -euo pipefail

LUXRAIL_REPO=${LUXRAIL_REPO:-LuxForge-ltd/luxrail}
UPSTREAM_REPO=${UPSTREAM_REPO:-musakkka/openrail}
UPSTREAM_BRANCH=${UPSTREAM_BRANCH:-main}
SYNC_ASSIGNEE=${SYNC_ASSIGNEE:-30-something-programmer}
SYNC_GIT_URL_BASE=${SYNC_GIT_URL_BASE:-https://github.com}
SYNC_DATE=${SYNC_DATE:-$(date -u +%Y-%m-%d)}
DEFAULT_TEST_CMD='make dist-placeholder && go vet ./... && go test ./... && (cd web && npm ci --no-audit --no-fund && npm test)'
SYNC_TEST_CMD=${SYNC_TEST_CMD:-$DEFAULT_TEST_CMD}
BRANCH="sync/upstream-${SYNC_DATE}"

log() { printf '%s sync-upstream: %s\n' "$(date -u +%Y-%m-%dT%H:%M:%SZ)" "$*"; }
die() { log "error: $*"; exit 1; }

WORK=${SYNC_WORKDIR:-$(mktemp -d)}
cleanup() { [ -n "${SYNC_WORKDIR:-}" ] || rm -rf "$WORK"; }
trap cleanup EXIT

# git reads the token through askpass, so it never appears in a URL, argv or the log.
if [ -n "${GH_TOKEN:-}" ]; then
	cat >"$WORK/askpass" <<'EOF'
#!/bin/sh
case "$1" in Username*) echo x-access-token ;; *) printf '%s\n' "$GH_TOKEN" ;; esac
EOF
	chmod 700 "$WORK/askpass"
	export GIT_ASKPASS="$WORK/askpass" GIT_TERMINAL_PROMPT=0
fi
export GIT_AUTHOR_NAME=LuxForge GIT_AUTHOR_EMAIL=ben@luxforge.dev
export GIT_COMMITTER_NAME=LuxForge GIT_COMMITTER_EMAIL=ben@luxforge.dev

command -v git >/dev/null || die "git is not installed"
command -v gh >/dev/null || die "gh is not installed"

log "cloning ${LUXRAIL_REPO} (Dev) and fetching ${UPSTREAM_REPO} ${UPSTREAM_BRANCH}"
git clone --quiet --filter=blob:none --branch Dev "${SYNC_GIT_URL_BASE}/${LUXRAIL_REPO}.git" "$WORK/repo"
cd "$WORK/repo"
git remote add upstream "${SYNC_GIT_URL_BASE}/${UPSTREAM_REPO}.git"
git fetch --quiet upstream "$UPSTREAM_BRANCH"
UP=$(git rev-parse FETCH_HEAD)
DEV=$(git rev-parse HEAD)
UP_SHORT=$(git rev-parse --short "$UP")

if git merge-base --is-ancestor "$UP" "$DEV"; then
	log "up to date: Dev already contains upstream ${UPSTREAM_BRANCH} ${UP_SHORT}"
	exit 0
fi

NEW=$(git rev-list --count "$DEV..$UP")
UP_TAG=$(git describe --tags --abbrev=0 "$UP" 2>/dev/null || echo "$UP_SHORT")
log "upstream has ${NEW} new commit(s) up to ${UP_SHORT} (${UP_TAG}); merging on ${BRANCH}"
git switch --quiet -c "$BRANCH"

LOG_LINES=$(git log --no-merges --format='- %h %s' "$DEV..$UP" | head -n 60)
MESSAGE="Merge upstream openrail ${UP_TAG} (${UP_SHORT}) into LuxRail."

pr_body() { # $1 = status paragraph
	printf '## What changed\n\nMerges upstream `%s` `%s` at %s (%s) into LuxRail: %s new commit(s).\n\n%s\n\n## Upstream commits\n\n%s\n' \
		"$UPSTREAM_REPO" "$UPSTREAM_BRANCH" "$UP_SHORT" "$UP_TAG" "$NEW" "$1" "$LOG_LINES"
}

open_pr() { # $1 title, $2 body, $3 extra flags
	local url
	# shellcheck disable=SC2086
	url=$(gh pr create --repo "$LUXRAIL_REPO" --base Dev --head "$BRANCH" --title "$1" --body "$2" $3)
	gh pr edit "$url" --repo "$LUXRAIL_REPO" --add-assignee "$SYNC_ASSIGNEE" >/dev/null 2>&1 || true
	log "opened $url"
}

if ! git merge --no-ff --no-edit -m "$MESSAGE" "$UP" >/dev/null 2>&1; then
	CONFLICTS=$(git diff --name-only --diff-filter=U)
	[ -n "$CONFLICTS" ] || die "merge failed without conflicts"
	log "conflicts in: $(echo "$CONFLICTS" | tr '\n' ' ')"
	git add -A
	git commit --quiet --no-edit -m "$MESSAGE Conflicts are left marked for resolution."
	git push --quiet -u origin "$BRANCH"
	open_pr "Merge upstream openrail ${UP_TAG} (needs resolution)" \
		"$(pr_body "$(printf 'The merge has conflicts. They are committed with their markers on this branch; resolve them here, run the tests, then mark this pull request ready.\n\nConflicting files:\n\n%s' "$(echo "$CONFLICTS" | sed 's/^/- `/; s/$/`/')")")" "--draft"
	exit 10
fi

log "merged cleanly; running tests"
if ! (bash -c "$SYNC_TEST_CMD") >"$WORK/tests.log" 2>&1; then
	log "tests failed; last lines:"
	tail -n 40 "$WORK/tests.log" || true
	git push --quiet -u origin "$BRANCH"
	open_pr "Merge upstream openrail ${UP_TAG} (tests failing)" \
		"$(pr_body "$(printf 'The merge is clean but the tests failed, so Dev was not changed.\n\n```\n%s\n```' "$(tail -n 40 "$WORK/tests.log")")")" ""
	exit 11
fi
log "tests passed"

git push --quiet -u origin "$BRANCH"
# A plain push only fast-forwards; if Dev moved since the clone it is refused and Dev stays as is.
if ! git push --quiet origin "HEAD:refs/heads/Dev" 2>"$WORK/push.err"; then
	log "Dev moved during the run; ${BRANCH} is pushed but Dev was not changed"
	open_pr "Merge upstream openrail ${UP_TAG}" \
		"$(pr_body 'The merge is clean and the tests passed, but Dev moved while the job ran, so it was not fast-forwarded. Merge this pull request instead.')" ""
	exit 12
fi
log "Dev fast-forwarded to $(git rev-parse --short HEAD)"

# The morning task: a pull request from Dev into main for the operator to approve.
EXISTING=$(gh pr list --repo "$LUXRAIL_REPO" --base main --head Dev --state open --json url --jq '.[0].url // ""')
if [ -n "$EXISTING" ]; then
	gh pr comment "$EXISTING" --repo "$LUXRAIL_REPO" \
		--body "Dev now includes upstream openrail ${UP_TAG} (${UP_SHORT}), merged ${SYNC_DATE}." >/dev/null
	log "updated $EXISTING"
else
	url=$(gh pr create --repo "$LUXRAIL_REPO" --base main --head Dev \
		--title "Release Dev to main" \
		--body "$(pr_body 'Dev now includes this upstream merge and passed the tests. Approve and merge this pull request to release Dev to main.')")
	gh pr edit "$url" --repo "$LUXRAIL_REPO" --add-assignee "$SYNC_ASSIGNEE" >/dev/null 2>&1 || true
	gh pr edit "$url" --repo "$LUXRAIL_REPO" --add-reviewer "$SYNC_ASSIGNEE" >/dev/null 2>&1 || true
	log "opened $url for approval"
fi
log "done"
