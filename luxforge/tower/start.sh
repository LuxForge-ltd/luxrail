#!/bin/sh
# Starts the trial Tower on LuxRail.
#
# The Tower serves its API and web frontend on loopback only unless it has its own certificate.
# LuxRail's Caddy terminates TLS for the service's domain and forwards plain HTTP to $PORT, so
# forward.py forwards $PORT to the Tower's loopback port. Tower sign-in still guards every page.
#
# Service variables it reads at run time:
#   TOWER_ADMIN_PASSWORD  sealed; sets the admin user's password on every start (optional)
#   TOWER_ADMIN_USER      the admin user's name (default admin)
#   PORT                  provided by LuxRail (default 8080)
#   TOWER_API_PORT        the Tower's loopback port (default 8480)
set -eu

root="${TOWER_ROOT:-/var/lib/tower}"
port="${PORT:-8080}"
api_port="${TOWER_API_PORT:-8480}"
run_as="setpriv --reuid=tower --regid=tower --init-groups --"

if [ "$api_port" = "$port" ]; then
  echo "PORT and TOWER_API_PORT must differ (both are $port)." >&2
  exit 2
fi

# A new LuxRail volume is owned by root; the Tower runs as tower (uid 10001).
if [ "$(id -u)" = 0 ]; then
  install -d -m 0700 "$root"
  chown -R tower:tower "$root"
else
  run_as=""
fi

# First start: the directory tree and one built-in job, so there is something to watch.
mark="$root/.luxrail"
$run_as sh -c "mkdir -p '$mark'"
if [ ! -e "$mark/init" ]; then
  $run_as tower init
  $run_as touch "$mark/init"
fi
if [ ! -e "$mark/host-metrics" ]; then
  $run_as tower builtin add host-metrics
  $run_as touch "$mark/host-metrics"
fi

if [ -n "${TOWER_ADMIN_PASSWORD:-}" ]; then
  printf '%s\n' "$TOWER_ADMIN_PASSWORD" \
    | $run_as tower user add "${TOWER_ADMIN_USER:-admin}" --role admin --password-stdin
else
  echo "TOWER_ADMIN_PASSWORD is not set: no user can sign in to the web frontend." >&2
fi
# LuxRail passes every service variable to the container; the Tower and its jobs need neither.
unset TOWER_ADMIN_PASSWORD LUXFORGE_FEED_TOKEN

$run_as /usr/local/bin/python3.12 /usr/local/lib/tower-luxrail/forward.py "$port" "$api_port" &

TOWER_API_HOST=127.0.0.1 TOWER_API_PORT="$api_port" exec $run_as tower run
