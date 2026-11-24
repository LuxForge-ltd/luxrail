#!/usr/bin/env bash
# Tests for sync-upstream.sh against local bare repositories and a fake `gh`.
# Run: bash luxforge/sync/test/test-sync-upstream.sh   (needs bash and git; no network)
# Each check's condition is a single-quoted string that `check` evaluates later, so variables in it
# (code, before, ...) are read there: SC2016 and SC2034 do not apply.
# shellcheck disable=SC2016,SC2034
set -euo pipefail

HERE=$(cd "$(dirname "$0")" && pwd)
SCRIPT="$HERE/../sync-upstream.sh"
ROOT=$(mktemp -d)
trap 'rm -rf "$ROOT"' EXIT
PASS=0
FAIL=0

export GIT_AUTHOR_NAME=Test GIT_AUTHOR_EMAIL=test@example.invalid
export GIT_COMMITTER_NAME=Test GIT_COMMITTER_EMAIL=test@example.invalid
export GIT_CONFIG_NOSYSTEM=1 HOME="$ROOT/home"
mkdir -p "$HOME" "$ROOT/bin"
git config --global init.defaultBranch main
git config --global protocol.file.allow always

# Fake gh: records each call, prints a URL for `pr create`, an empty list for `pr list`.
cat >"$ROOT/bin/gh" <<'EOF'
#!/usr/bin/env bash
printf '%s\n' "$*" >>"$GH_LOG"
case "$1 $2" in
  "pr create") echo "https://github.invalid/pr/1" ;;
  "pr list") echo "${GH_EXISTING_PR:-}" ;;
esac
EOF
chmod +x "$ROOT/bin/gh"
export PATH="$ROOT/bin:$PATH"

setup() { # fresh upstream and LuxRail remotes; LuxRail Dev = upstream main + one LuxForge commit
	rm -rf "$ROOT/remotes" "$ROOT/seed" "$ROOT/gh.log"
	mkdir -p "$ROOT/remotes/musakkka" "$ROOT/remotes/LuxForge-ltd"
	export GH_LOG="$ROOT/gh.log"
	: >"$GH_LOG"
	git init -q "$ROOT/seed"
	(
		cd "$ROOT/seed"
		printf 'line one\nline two\n' >shared.txt
		git add . && git commit -qm "upstream start"
		git clone -q --bare . "$ROOT/remotes/musakkka/openrail.git"
		git clone -q --bare . "$ROOT/remotes/LuxForge-ltd/luxrail.git"
		git push -q "$ROOT/remotes/LuxForge-ltd/luxrail.git" main:Dev
		git switch -q -c dev
		echo luxforge >LUXRAIL.md && git add . && git commit -qm "luxforge change"
		git push -q "$ROOT/remotes/LuxForge-ltd/luxrail.git" dev:Dev
	)
}

upstream_commit() { # $1 file, $2 content
	(
		cd "$ROOT/seed"
		git switch -q main
		printf '%s\n' "$2" >"$1"
		git add . && git commit -qm "upstream: change $1"
		git push -q "$ROOT/remotes/musakkka/openrail.git" main
	)
}

run_sync() { # prints the exit code; output goes to $ROOT/out.log
	set +e
	SYNC_GIT_URL_BASE="file://$ROOT/remotes" SYNC_DATE=2026-11-27 GH_TOKEN=test-token \
		SYNC_TEST_CMD="${TEST_CMD:-true}" bash "$SCRIPT" >"$ROOT/out.log" 2>&1
	echo $?
	set -e
}

dev_rev() { git --git-dir="$ROOT/remotes/LuxForge-ltd/luxrail.git" rev-parse Dev; }
branch_exists() { git --git-dir="$ROOT/remotes/LuxForge-ltd/luxrail.git" rev-parse -q --verify "refs/heads/$1" >/dev/null; }

check() { # $1 name, $2 condition (eval'd)
	if eval "$2"; then
		PASS=$((PASS + 1)); echo "ok   $1"
	else
		FAIL=$((FAIL + 1)); echo "FAIL $1"; sed 's/^/     | /' "$ROOT/out.log"
	fi
}

# 1. Nothing new upstream: exit 0, no branch, no pull request.
setup
before=$(dev_rev)
code=$(run_sync)
check "up to date exits 0" '[ "$code" = 0 ]'
check "up to date leaves Dev" '[ "$(dev_rev)" = "$before" ]'
check "up to date opens no PR" '! grep -q "pr create" "$GH_LOG"'
check "up to date says so" 'grep -q "up to date" "$ROOT/out.log"'

# 2. Clean merge, tests pass: Dev fast-forwarded, branch pushed, Dev -> main PR opened.
setup
upstream_commit new.txt "new upstream file"
before=$(dev_rev)
code=$(run_sync)
check "clean merge exits 0" '[ "$code" = 0 ]'
check "clean merge moves Dev" '[ "$(dev_rev)" != "$before" ]'
check "Dev keeps the LuxForge change" 'git --git-dir="$ROOT/remotes/LuxForge-ltd/luxrail.git" cat-file -e Dev:LUXRAIL.md'
check "Dev has the upstream change" 'git --git-dir="$ROOT/remotes/LuxForge-ltd/luxrail.git" cat-file -e Dev:new.txt'
check "Dev is a fast-forward of the old Dev" 'git --git-dir="$ROOT/remotes/LuxForge-ltd/luxrail.git" merge-base --is-ancestor "$before" Dev'
check "sync branch pushed" 'branch_exists sync/upstream-2026-11-27'
check "Dev -> main PR opened" 'grep -q "pr create --repo LuxForge-ltd/luxrail --base main --head Dev" "$GH_LOG"'
check "merge authored as LuxForge" '[ "$(git --git-dir="$ROOT/remotes/LuxForge-ltd/luxrail.git" log -1 --format=%an Dev)" = LuxForge ]'
check "token never printed" '! grep -q test-token "$ROOT/out.log"'

# 2b. A Dev -> main PR already open: comment on it instead of opening another.
setup
upstream_commit new.txt "new upstream file"
code=$(GH_EXISTING_PR=https://github.invalid/pr/9 run_sync)
check "existing release PR gets a comment" 'grep -q "pr comment https://github.invalid/pr/9" "$GH_LOG"'
check "no second release PR" '! grep -q -- "pr create .*--base main" "$GH_LOG"'

# 3. Conflict: exit 10, draft PR into Dev, Dev untouched.
setup
(cd "$ROOT/seed" && git switch -q dev && printf 'luxforge edit\nline two\n' >shared.txt && git commit -qam "luxforge edits shared" && git push -q "$ROOT/remotes/LuxForge-ltd/luxrail.git" dev:Dev)
upstream_commit shared.txt "upstream edit"
before=$(dev_rev)
code=$(run_sync)
check "conflict exits 10" '[ "$code" = 10 ]'
check "conflict leaves Dev" '[ "$(dev_rev)" = "$before" ]'
check "conflict pushes the branch" 'branch_exists sync/upstream-2026-11-27'
check "conflict opens a PR into Dev" 'grep -q "pr create --repo LuxForge-ltd/luxrail --base Dev" "$GH_LOG"'
check "conflict PR is a draft" 'grep -q -- "--draft" "$GH_LOG"'
check "conflict PR names the file" 'grep -q "shared.txt" "$GH_LOG"'
check "conflict markers committed on the branch" 'git --git-dir="$ROOT/remotes/LuxForge-ltd/luxrail.git" show sync/upstream-2026-11-27:shared.txt | grep -q "^<<<<<<<"'

# 4. Tests fail: exit 11, PR into Dev, Dev untouched.
setup
upstream_commit new.txt "new upstream file"
before=$(dev_rev)
code=$(TEST_CMD='echo boom; exit 3' run_sync)
check "failing tests exit 11" '[ "$code" = 11 ]'
check "failing tests leave Dev" '[ "$(dev_rev)" = "$before" ]'
check "failing tests open a PR into Dev" 'grep -q "pr create --repo LuxForge-ltd/luxrail --base Dev" "$GH_LOG"'
check "failing tests show the output" 'grep -q boom "$ROOT/out.log"'

# 5. Dev moves while the tests run: exit 12, Dev keeps the other change, nothing forced.
setup
upstream_commit new.txt "new upstream file"
MOVE="cd '$ROOT/seed' && git switch -q dev && echo other >other.txt && git add . && git commit -qm other && git push -q '$ROOT/remotes/LuxForge-ltd/luxrail.git' dev:Dev"
code=$(TEST_CMD="$MOVE" run_sync)
check "Dev moved exits 12" '[ "$code" = 12 ]'
check "Dev keeps the concurrent change" 'git --git-dir="$ROOT/remotes/LuxForge-ltd/luxrail.git" cat-file -e Dev:other.txt'
check "Dev moved opens a PR into Dev" 'grep -q "pr create --repo LuxForge-ltd/luxrail --base Dev" "$GH_LOG"'

echo
echo "$PASS passed, $FAIL failed"
[ "$FAIL" = 0 ]
