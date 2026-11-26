# Trial Tower

A Tower (LuxForge's control plane for contracted jobs, from `LuxForge-ltd/tower`) running as a
LuxRail service, so its behaviour can be watched on a real install: its web frontend on a LuxRail
domain, its logs in the console, its state on a volume. It is a trial, not a LuxHome element: it
manages nothing but its own built-in jobs.

| File | Does |
| --- | --- |
| `Dockerfile` | Python 3.12 and tini, with `luxforge-tower[postgres]==$TOWER_VERSION` installed from the Foundry feed into `/opt/tower` |
| `start.sh` | Gives the volume to the `tower` user, creates the Tower directory and the `host-metrics` job on first start, sets the admin password, starts the forwarder, then runs `tower run` as `tower` |
| `forward.py` | Forwards `$PORT` to the Tower's loopback API (see below) |
| `luxrail.json` | Dockerfile build, ready check on `/api/v1/ready`, restart on failure (10 tries), one replica |

## Why a forwarder

The Tower only serves plain HTTP on loopback; anywhere else it insists on its own certificate.
LuxRail's Caddy terminates TLS for the domain and connects to the container over plain HTTP, so
`forward.py` copies bytes between `0.0.0.0:$PORT` and `127.0.0.1:8480`. HTTP, server-sent events
and the WebSocket pass through unchanged; the WebSocket's same-origin check holds because Caddy
keeps the `Host` header. Every page still needs a Tower sign-in. The session cookie is not marked
`Secure`, because the Tower itself sees plain HTTP; browsers still send it only to this domain.

## Deploying it (once)

1. **Service.** In the console, in `luxrail-ops`, add a service from the GitHub repository
   `LuxForge-ltd/luxrail`, branch `Dev`, with **Config file path** `luxforge/tower/luxrail.json`.
2. **Variables.**
   - `LUXFORGE_FEED_TOKEN`, **sealed**: a read-only Foundry feed token (Packaging, Read). The build
     mounts it as a BuildKit secret; it never reaches an image layer, and `start.sh` removes it
     from the Tower's environment.
   - `TOWER_ADMIN_PASSWORD`, **sealed**: the password of the Tower user `admin` (or
     `TOWER_ADMIN_USER`). It is set again on every start, so changing the variable and
     redeploying changes the password. Without it, nobody can sign in to the frontend.
   - `TOWER_VERSION` (optional): the release to install; the Dockerfile's default is `0.0.54`. The
     web frontend is in the wheel from 0.0.49. The version must be on the feed.
3. **Volume.** Attach a volume at `/var/lib/tower`. Without it, every deploy starts an empty Tower.
4. **Domain.** Give the service a domain; LuxRail sets `PORT`. Optionally switch on **This is a web
   frontend** and set the Website to that domain.
5. **Deploy.** The build log names `LUXFORGE_FEED_TOKEN` under "mounted as build secrets" and
   ends with `tower --version`. The first start logs `Tower directory ready`, `wrote
   .../contracts/host-metrics.json`, `user admin saved` and `forwarding 0.0.0.0:<port>`.

## Watching it

- Sign in at the domain as `admin`. The Jobs page shows `host-metrics` (every 60 s: CPU, load,
  memory, uptime and disk of the VM as the container sees it).
- The service's logs carry the Tower's own log (JSON lines from Foundry's logger), job runs and the
  audit trail (sign-ins, control actions).
- More jobs: `tower builtin list`, then `tower builtin add <name>` as user `tower` in the service's
  shell, or contracts through the frontend or the API.

## Upgrading

Set `TOWER_VERSION` to the new release and redeploy. The volume carries the Tower's state across
versions. A release that adds its own listener, such as a status page on another port, is
reached the same way: point `PORT` (or a second `forward.py`) at it in `start.sh`.

## Stopping it

`tower disable` (as user `tower` in the service's shell) makes `tower run` exit 0, and with
`ON_FAILURE` LuxRail leaves it down. The flag is on the volume, so a later start exits 4 ("started
while disabled") until `tower enable` is run; the Tower's exit codes 2, 4 and 6 mean "do not just
restart", which is why LuxRail stops retrying after 10 tries. Remove the service and its volume to
end the trial.
