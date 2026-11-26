---
title: Set up the CLI
nav: The CLI
description: Install the LuxRail CLI on your laptop or in CI, log in to your console, and optionally make it answer to railway.
---

Install the `luxrail` CLI on your laptop or in CI, log in to your console, and optionally make it answer to `railway`.

## Install

On macOS or Linux:

```sh
curl -fsSL https://github.com/LuxForge-ltd/luxrail/releases/latest/download/install.sh | sh -s -- --cli
```

It installs only the `luxrail` command, no server parts, and needs no root. It goes to `/usr/local/bin` when you can write there (or run it as root), otherwise to `~/.local/bin`. If that folder is not on your `PATH`, the installer says so; add it in your shell's profile:

```sh
export PATH="$HOME/.local/bin:$PATH"
```

With wget:

```sh
wget -qO- https://github.com/LuxForge-ltd/luxrail/releases/latest/download/install.sh | sh -s -- --cli
```

Archives for every platform are on the repository's Releases page.

On Windows, take `luxrail.exe` from the `windows_amd64` archive (or build it with `make cross`, which writes `bin/windows_amd64/luxrail.exe`) and put it in a folder on your user `PATH`. With Smart App Control on, Windows only runs it if it is signed by a certificate Microsoft trusts.

## Log in

```sh
luxrail login https://console.example.com
```

The CLI prints a code and opens the console's **Approve CLI login** page in your browser. Check that the code matches and approve it. On a machine with no browser, add `--browserless` and open the printed link anywhere you are logged in.

Check it worked:

```sh
luxrail whoami
```

The CLI keeps one token per console in `~/.config/luxrail/config.json` (mode 0600; on Windows `%USERPROFILE%\.config\luxrail\config.json`), so you can log in to several installs.

### A console with its own CA

An install on a LAN often serves its console with a certificate from its own CA, such as Caddy's internal CA. Give the CLI that CA's certificate when you log in:

```sh
luxrail login https://10.0.0.5 --ca-file caddy-root.crt
```

The CLI trusts that CA, on top of the system's roots, for this console only, and keeps a copy of the certificate with the login, so later commands need nothing more. TLS verification stays on. The file must hold CA certificates only; a file with a private key is refused. For a single run, or in CI, set `LUXRAIL_CA_FILE` to the file instead. Without it, the CLI says the certificate is signed by a CA it does not trust and suggests `--ca-file`. `luxrail logout` forgets the CA with the token.

On Windows, Go programs read only the Windows certificate store, so `SSL_CERT_FILE` has no effect; use `--ca-file`.

## Tokens for servers and CI

Create a token in the console. There are two kinds:

- **Settings**, **API tokens**: an account token. It acts as you, with your role.
- A project's **Project Settings**, **Tokens**, under **Environment tokens**: a token that can deploy, read logs and set variables in one environment. It suits CI best.

Copy the token at once; it is not shown again.

Then either log in with it:

```sh
luxrail login https://console.example.com --token 5050_xxxxxxxx
```

or skip login and set two variables, which suits CI best:

```sh
export LUXRAIL_URL=https://console.example.com
export LUXRAIL_TOKEN=5050_xxxxxxxx
```

A GitHub Actions step that deploys the checked-out code to a linked service:

```yaml
- name: Deploy to luxrail
  env:
    LUXRAIL_URL: ${{ secrets.LUXRAIL_URL }}
    LUXRAIL_TOKEN: ${{ secrets.LUXRAIL_TOKEN }}
  run: |
    curl -fsSL https://github.com/LuxForge-ltd/luxrail/releases/latest/download/install.sh | sh -s -- --cli
    export PATH="$HOME/.local/bin:$PATH"
    luxrail up -c --service web
```

`-c` (`--ci`) streams the build logs and exits when the deployment finishes, with a non-zero code if it failed.

## Link a project

In your app's folder:

```sh
luxrail link
luxrail up
```

`luxrail link` asks for the project, environment and service, and remembers them for this folder. `luxrail up` uploads the folder, builds it and deploys it.

## The railway alias

```sh
luxrail alias install
```

This adds a `railway` command next to `luxrail`, unless a real `railway` binary is already on your `PATH`. The first time you run it, it prints `Using luxrail at <url>; run luxrail alias remove to undo`. As `railway`, it accepts Railway's verbs and flags and reads `RAILWAY_TOKEN` and `RAILWAY_API_TOKEN`; put a luxrail token in them, not a Railway one. `luxrail alias remove` takes it away. See [The railway alias](../cli/railway-alias.md).

## Upgrade the CLI

Run the install command again. It replaces the binary in place.

## What success looks like

- `luxrail whoami` prints your email and the console's address.
- `luxrail status` (or `railway status` with the alias) prints the linked project, environment and service.

## Next steps

- [Deploy with the CLI](deploy-with-cli.md)
- [CLI overview](../cli/overview.md)
- [Global options](../cli/global-options.md)
