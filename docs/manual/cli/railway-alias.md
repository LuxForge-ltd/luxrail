---
title: The railway alias
description: The `luxrail` binary also answers to `railway`, so scripts and habits built on Railway's CLI keep working.
---

The `luxrail` binary also answers to `railway`, so scripts and habits built on Railway's CLI keep working.

## How it works

`luxrail alias install` adds a `railway` symlink next to the `luxrail` binary. It leaves a real `railway` binary alone if one is already on your `PATH`. When the binary runs as `railway`, it accepts Railway's verbs and flags, for example:

```sh
railway login
railway link
railway up --service api
railway variables --set "FOO=bar"
railway add -d postgres
railway logs
railway run -- npm run migrate
```

The first time, it prints one line: `Using luxrail at <url>; run luxrail alias remove to undo`.

A verb LuxRail does not have answers with `railway <verb> is not available in LuxRail. See https://github.com/LuxForge-ltd/luxrail/blob/main/docs/manual/coming-from-railway.md` and exits non-zero.

## Manage the alias

```sh
luxrail alias install
luxrail alias remove
```

## Tokens

As `railway`, the CLI also reads `RAILWAY_TOKEN` and `RAILWAY_API_TOKEN`. Use a LuxRail API token as the value, not a Railway token.

## Next steps

- [Coming from Railway](../coming-from-railway.md)
- [CLI overview](overview.md)
