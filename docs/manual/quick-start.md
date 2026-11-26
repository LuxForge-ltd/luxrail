---
title: Quick start
description: Install LuxRail on a Linux VM, open the console, deploy a GitHub repo and add Postgres.
---

Install LuxRail on a Linux VM, open the console, deploy a GitHub repo and add Postgres.

## What you need

- A VM with a public IPv4 address and root access. Ubuntu 24.04 is the tested OS (amd64 or arm64). 2 vCPU and 4 GB of RAM is a comfortable start.
- Ports 80 and 443 open to the internet in your cloud provider's firewall. See [Firewall](guides/firewall.md).
- No domain. Without one, the console and your services get `sslip.io` names with real certificates.

## Install

On the VM:

```sh
curl -fsSL https://github.com/LuxForge-ltd/luxrail/releases/latest/download/install.sh | sudo sh
```

With wget:

```sh
wget -qO- https://github.com/LuxForge-ltd/luxrail/releases/latest/download/install.sh | sudo sh
```

Piped like this, the script asks no questions and uses safe defaults. For a guided setup that asks for your domain and email, download it first and run it in a terminal:

```sh
curl -fsSLo install.sh https://github.com/LuxForge-ltd/luxrail/releases/latest/download/install.sh && sudo sh install.sh
```

Every answer is also a flag or a `FIFTY_*` environment variable; see [Install](guides/install.md) and [Install options](reference/install-options.md).

The installer sets up Docker, the control plane (`5050d`), the agent, Caddy for HTTPS and a private registry. It finishes with a one-time setup link:

```text
LuxRail.0 is ready.

  Open this link to create your account (valid 24 hours, works once):
  https://console-ab12.203-0-113-10.sslip.io/setup?token=...
```

> **Note:** The recovery key is written to `/etc/5050/recovery-key.txt` (mode 0600). Copy it into a password manager and delete the file. It decrypts your backups and state snapshots.

## Create your account

1. Open the printed link in your browser.
2. Create the owner account: name, email and a password of at least 10 characters.
3. You land on an empty dashboard.

The setup link works once and expires after 24 hours. `sudo 5050d setup-token` on the host prints a new one.

## Deploy a repo

1. On the dashboard, select **New project**, then **GitHub Repository**.
2. Paste a public repo address, for example `https://github.com/LuxForge-ltd/luxrail`, and set **Root directory** to `examples/node-hello`. Private repos need the [GitHub App](guides/github-app.md).
3. Select **Deploy Now**.
4. The service card shows **Building**, then **Deploying**, then **Active**.
5. Open the service, go to **Settings**, **Networking** and select **Generate Domain**. Open the address.

![The Networking settings of a web service: a generated domain and a custom domain, each with a valid certificate, the Generate Domain, Custom Domain and TCP Proxy buttons, and the private domain web.internal](shot:networking "Settings, Networking: the generated name and a custom domain, each with a valid certificate.")

## Add Postgres

1. On the canvas, select **+ Add**, then **Database**, then **Postgres**.
2. Open your app service, go to **Variables** and add `DATABASE_URL` with the value `${{Postgres.DATABASE_URL}}`.
3. Select **Deploy** in the banner at the top of the canvas to apply the staged change.

![The Variables tab of the api service: DATABASE_URL open for editing with the value ${{Postgres.DATABASE_URL}}, above REDIS_URL and the other variables](shot:variables "A reference to Postgres in the Variables tab, written the way Railway writes it.")

The canvas now draws a line from your app to Postgres, the dashed lines below.

![The project canvas: web, api with 3 replicas, a cron job, a worker, Postgres grouped with PgBouncer, Redis, docs and search, joined by dashed reference lines](shot:canvas "A larger project on the canvas. Each dashed line is a variable reference.")

## What success looks like

- The generated address opens your app over HTTPS with no certificate warning.
- The app's **Variables** tab shows `DATABASE_URL` resolved to the Postgres service.

## Next steps

- [Your domains](guides/console-domain.md) and [Custom domains and HTTPS](guides/custom-domains.md).
- [Backups bucket](guides/backups-bucket.md): turn on database backups.
- [Set up the CLI](guides/cli-setup.md): `luxrail up` from your laptop.
- [Production checklist](production-checklist.md).
