---
title: Templates and the Deploy on LuxRail button
description: Deploy ready-made apps such as n8n, Umami or Uptime Kuma in one click, and add a Deploy on LuxRail button to your own README.
---

Deploy ready-made apps such as n8n, Umami or Uptime Kuma in one click, and add a Deploy on LuxRail button to your own README.

## Deploy a template

1. On the canvas, select **+ Add**, then **Template**.
2. Pick a template. The detail page shows the services it creates and its inputs; generated values (secrets) are prefilled and can be regenerated.
3. Fill in required inputs and select **Deploy**.

Templates deploy their services in dependency order, for example Postgres before n8n. From the CLI: `luxrail deploy --template n8n`.

The list is in [Templates reference](../reference/templates.md).

## Deep links

`https://<your console>/new/template/<id>?KEY=value` opens a template with inputs prefilled.

## Writing a template

See [Writing a template](../contribute/writing-a-template.md).
