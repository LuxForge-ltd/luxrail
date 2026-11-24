# Weekly upstream sync

`sync-upstream.sh` merges upstream openrail (`musakkka/openrail`, `main`) into LuxRail's `Dev`
every Friday at 02:00 UTC. It runs as a cron service on LuxForge's own OpenRail/LuxRail install
(VM 105 on Savathun), so its runs, logs and failures are visible in the console.

## What a run does

1. Clones LuxRail (`Dev`) and fetches upstream `main`.
2. **Nothing new:** logs "up to date" and exits 0.
3. Branches `sync/upstream-<date>` from `Dev` and merges upstream into it, as LuxForge.
4. **Conflicts:** commits the merge with its conflict markers, pushes the branch and opens a
   **draft** pull request into `Dev` titled "Merge upstream openrail … (needs resolution)" that lists
   the files. `Dev` is not changed. Exit 10.
5. **Clean merge:** runs the tests (`make dist-placeholder`, `go vet ./...`, `go test ./...`, then
   `npm ci` and `npm test` in `web/`). If they fail, it pushes the branch and opens a pull request
   into `Dev` with the failure. `Dev` is not changed. Exit 11.
6. **Tests pass:** pushes the branch, then fast-forwards `Dev` to it. The push is never forced; if
   `Dev` moved during the run, the push is refused, a pull request into `Dev` is opened instead and
   the run exits 12.
7. Makes sure a pull request **"Release Dev to main"** is open and assigned to the operator. That
   is the morning task: approving and merging it releases `Dev` to `main`. If one is already open,
   the run adds a comment to it instead.

The token reaches git through an askpass helper only, so it never appears in a URL, the process
list or the log.

## Deploying it (once)

1. **Token.** Create a fine-grained GitHub token for `LuxForge-ltd/luxrail` only, with
   *Contents*, *Pull requests* and *Workflows* set to *Read and write* (upstream merges can touch
   `.github/workflows/`, which GitHub refuses to push without it; Actions stays off, so nothing
   runs), and an expiry. Keep it in the KeePassXC vault as `luxrail/github-sync-token`.
2. **Service.** In the console, in a project such as `luxrail-ops`, add a service from the GitHub
   repository `LuxForge-ltd/luxrail`, branch `Dev`, with the config file
   `luxforge/sync/railway.json`. That file builds `luxforge/sync/Dockerfile` from the repository
   root.
3. **Schedule.** In the service's *Settings > Deploy*, set *Cron Schedule* to `0 2 * * 5` (UTC)
   and the restart policy to *Never*. These are set on the service, not in `railway.json`:
   openrail v0.1.9's scheduler and *Run now* only read a service's own schedule, and it refuses a
   schedule on the service while the config file sets one. (A LuxRail fix will let the file carry
   it again.)
4. **Variables.** Add `GH_TOKEN` as a **sealed** variable (paste it from the vault; it is never
   shown again). Optional: `SYNC_ASSIGNEE` (default `30-something-programmer`), `SYNC_TEST_CMD`.
5. **Deploy**, then use **Run now** once and read the run's log. A first run with nothing new
   upstream ends with "up to date".

On joyeuse, `L:\LuxForge\bin\luxrail-deploy-sync.ps1` does steps 2 to 5 through the OpenRail API.
It asks once for the vault password and is safe to run again.

## Reading the result

| Exit | Meaning | What to do |
| --- | --- | --- |
| 0 | Up to date, or merged into `Dev` | Approve "Release Dev to main" when ready |
| 10 | Conflicts | Resolve them on the draft pull request's branch, run the tests, mark it ready |
| 11 | Tests failed | Fix on the pull request's branch |
| 12 | `Dev` moved during the run | Merge the opened pull request into `Dev` |
| other | The job itself failed | Read the run's log |

## Tests

```
bash luxforge/sync/test/test-sync-upstream.sh
```

They use local bare repositories and a fake `gh`, so they need bash and git but no network or
token. They cover: up to date, clean merge (Dev fast-forwarded, release pull request opened or
commented), conflicts (draft pull request, markers committed, Dev untouched), failing tests, and
`Dev` moving during the run.
