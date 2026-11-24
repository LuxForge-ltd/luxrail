# The upstream-sync cron service: git, gh, Go and Node, matching upstream's CI (Go from go.mod,
# Node 24), so the job can merge upstream openrail into LuxRail and run the tests.
# Build context: the repository root (see railway.json next to this file).
FROM node:24-bookworm-slim AS node

FROM golang:1.26-bookworm
COPY --from=node /usr/local/bin/node /usr/local/bin/node
COPY --from=node /usr/local/lib/node_modules /usr/local/lib/node_modules
RUN ln -s ../lib/node_modules/npm/bin/npm-cli.js /usr/local/bin/npm \
 && ln -s ../lib/node_modules/npm/bin/npx-cli.js /usr/local/bin/npx \
 && apt-get update \
 && apt-get install -y --no-install-recommends git make ca-certificates curl gnupg \
 && install -d -m 0755 /etc/apt/keyrings \
 && curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg -o /etc/apt/keyrings/githubcli-archive-keyring.gpg \
 && chmod go+r /etc/apt/keyrings/githubcli-archive-keyring.gpg \
 && echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" > /etc/apt/sources.list.d/github-cli.list \
 && apt-get update \
 && apt-get install -y --no-install-recommends gh \
 && rm -rf /var/lib/apt/lists/* \
 && useradd --create-home --uid 10001 luxrail
COPY luxforge/sync/sync-upstream.sh /usr/local/bin/sync-upstream
RUN chmod 0755 /usr/local/bin/sync-upstream
USER luxrail
ENV GOPATH=/home/luxrail/go GOCACHE=/home/luxrail/.cache/go-build
WORKDIR /home/luxrail
CMD ["/usr/local/bin/sync-upstream"]
