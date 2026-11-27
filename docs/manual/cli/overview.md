---
title: CLI overview
description: The `luxrail` CLI deploys, inspects and manages your services from a terminal or CI, and also answers to `railway`.
---

The `luxrail` CLI deploys, inspects and manages your services from a terminal or CI, and also answers to `railway`.

## Install

The CLI installs on macOS, Linux and Windows (Windows builds are experimental). On macOS or Linux:

```sh
curl -fsSL https://github.com/LuxForge-ltd/luxrail/releases/latest/download/install.sh | sh -s -- --cli
```

It goes to `/usr/local/bin` when that is writable (or you run it as root), else to `~/.local/bin`. Run `luxrail alias install` to also answer to `railway`. Release archives for every platform are on the GitHub Releases page.

## Log in and deploy

```sh
luxrail login https://console.example.com
cd my-app
luxrail init -n my-app
luxrail up --new
```

## Commands

| Command | What it does |
|---|---|
| [login](login.md), [logout](logout.md), [whoami](whoami.md) | Manage the stored token |
| [init](init.md), [link](link.md), [unlink](unlink.md) | Create and link projects |
| [status](status.md), [list](list.md), [open](open.md) | See where you are |
| [up](up.md) | Upload this directory and deploy it |
| [redeploy](redeploy.md), [restart](restart.md), [down](down.md) | Act on the active deployment |
| [deploy](deploy.md), [templates](templates.md) | Deploy templates |
| [add](add.md) | Add a database, repo, image or empty service |
| [import](import.md) | Import from Railway, Docker Compose or Heroku |
| [logs](logs.md) | Deploy, build and HTTP logs |
| [variables](variables.md) | List, set and delete variables |
| [run](run.md), [shell](shell.md) | Run local commands with the service's variables |
| [ssh](ssh.md) | Shell inside a running replica |
| [connect](connect.md) | Database clients through a tunnel |
| [domain](domain.md), [tcp-proxy](tcp-proxy.md) | Public endpoints |
| [environment](environment.md), [service](service.md) | Switch and change environments and services |
| [scale](scale.md), [volume](volume.md), [cron](cron.md) | Replicas, volumes and cron runs |
| [metrics](metrics.md), [usage](usage.md) | Resource use |
| [config](config.md) | Export settings to a file |
| [host](host.md), [upgrade](upgrade.md) | Admin: hosts and upgrades |
| [alias](alias.md), [completion](completion.md), [version](version.md), [self-update](../guides/cli-setup.md#upgrade-the-cli) | The CLI itself |

## Next steps

- [Global options](global-options.md)
- [The railway alias](railway-alias.md)
