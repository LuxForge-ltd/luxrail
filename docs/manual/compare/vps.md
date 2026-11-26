---
title: LuxRail and a plain VPS
description: Running apps on a VPS by hand is cheapest in software but costs time; LuxRail automates the parts that take it.
---

Running apps on a VPS by hand is cheapest in software but costs time; LuxRail automates the parts that take it.

## What you would do by hand

- Install Docker, a reverse proxy and certificates.
- Write deploy scripts, health checks and rollbacks.
- Set up database backups, test restores and watch disk use.
- Wire logs and metrics.
- Repeat it on every new server and keep it patched.

## What LuxRail does

- One command installs it all with HTTPS by default.
- Git deploys with health checks, zero-downtime rollouts for services without a volume, and rollbacks that restore variables.
- Databases with point-in-time backups to your bucket and monthly restore tests.
- Logs, metrics, monitors and alerts.
- New hosts join with one command.
- Unattended security updates and self-reverting upgrades.

## What stays yours

The VM, the OS reboots, the bucket and the bill. See [Honest limits](../reference/honest-limits.md).
