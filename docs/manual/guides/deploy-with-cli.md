---
title: Deploy with the CLI
description: `luxrail up` uploads the current directory, builds it on your hosts and streams the logs until the deploy is live.
---

`luxrail up` uploads the current directory, builds it on your hosts and streams the logs until the deploy is live.

## Steps

1. Install the CLI: `curl -fsSL https://github.com/LuxForge-ltd/luxrail/releases/latest/download/install.sh | sh -s -- --cli`.
2. Log in: `luxrail login https://console.example.com`.
3. In your app's directory, create or link a project: `luxrail init -n my-app` or `luxrail link`.
4. Deploy: `luxrail up --new` the first time, `luxrail up` afterwards.

```
$ luxrail up
Uploaded 3.2 MB
... build log ...
... first deploy log lines ...
Deployed: https://my-app-k3x9.203-0-113-10.sslip.io
```

## What is uploaded

The directory is packed as a gzipped tar, skipping `.git` and files matched by `.gitignore` and `.5050ignore`. Uploads over 500 MB are refused. `--no-gitignore` includes ignored files; `--path-as-root PATH` uploads only `PATH` as the root.

## From CI

```sh
export LUXRAIL_URL=https://console.example.com
export LUXRAIL_TOKEN=5050_xxxxxxxx
luxrail up -c -s api -e production -m "$GIT_COMMIT_MESSAGE"
```

`-c` streams the build log and exits non-zero if the build or deploy fails. For CI, create an environment token in the project's **Project Settings**, **Tokens**. See [Set up the CLI](cli-setup.md#tokens-for-servers-and-ci).

## Next steps

- [CLI reference: up](../cli/up.md)
- [Config as code](config-as-code.md)
