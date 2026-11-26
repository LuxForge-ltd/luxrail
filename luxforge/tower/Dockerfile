# A trial Tower (LuxForge's job control plane) running as a LuxRail service, so its behaviour can
# be watched on a real install. Build context: the repository root (see luxrail.json next to
# this file). The Tower comes from LuxForge's private Foundry feed, pinned by TOWER_VERSION; set
# a service variable of that name to move to another release and redeploy.
#
# Service variables it reads at build time:
#   LUXFORGE_FEED_TOKEN  sealed; read token for the Foundry feed, mounted as a build secret only
#   TOWER_VERSION        optional; the luxforge-tower release to install
ARG PYTHON_IMAGE=python:3.12-slim-bookworm
FROM ${PYTHON_IMAGE}
ARG TOWER_VERSION=0.0.54

# tini reaps the children of job processes and passes SIGTERM to the Tower.
RUN apt-get update \
 && apt-get install -y --no-install-recommends tini \
 && rm -rf /var/lib/apt/lists/* \
 && pip install --no-cache-dir --root-user-action=ignore "uv==0.11.*" \
 && useradd --system --uid 10001 --home-dir /var/lib/tower --shell /usr/sbin/nologin tower \
 && install -d -m 0700 -o tower -g tower /var/lib/tower

# luxforge-tower and luxforge come only from the feed: uv takes each package from the first
# index that has it, and the named index comes before PyPI.
ENV UV_COMPILE_BYTECODE=1 UV_LINK_MODE=copy
RUN --mount=type=secret,id=LUXFORGE_FEED_TOKEN \
    test -s /run/secrets/LUXFORGE_FEED_TOKEN \
      || { echo "Set the sealed service variable LUXFORGE_FEED_TOKEN (Foundry feed read token)." >&2; exit 1; } \
 && uv venv /opt/tower --python /usr/local/bin/python3.12 \
 && UV_INDEX_FOUNDRY_USERNAME=feed UV_INDEX_FOUNDRY_PASSWORD="$(cat /run/secrets/LUXFORGE_FEED_TOKEN)" \
    uv pip install --python /opt/tower/bin/python --no-cache \
      --index "foundry=https://pkgs.dev.azure.com/LuxForge/4bcd7076-0dfa-4d4d-b42a-ffb92ec8ad9d/_packaging/Foundry/pypi/simple/" \
      "luxforge-tower[postgres]==${TOWER_VERSION}" \
 && /opt/tower/bin/tower --version

# forward.py carries LuxRail's plain-HTTP port to the Tower's loopback API (start.sh says why).
COPY luxforge/tower/start.sh /usr/local/bin/tower-luxrail
COPY luxforge/tower/forward.py /usr/local/lib/tower-luxrail/forward.py
RUN chmod 0755 /usr/local/bin/tower-luxrail

ENV PATH=/opt/tower/bin:$PATH TOWER_ROOT=/var/lib/tower TOWER_PYTHON_3_12=/usr/local/bin/python3.12
WORKDIR /var/lib/tower
STOPSIGNAL SIGTERM
# Starts as root only to give the volume to the tower user; start.sh drops to it before anything
# else runs.
ENTRYPOINT ["/usr/bin/tini", "--", "/usr/local/bin/tower-luxrail"]
