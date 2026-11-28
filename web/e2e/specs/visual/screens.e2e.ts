// SPDX-License-Identifier: AGPL-3.0-only

// Screens for visual comparison. Runs only when E2E_SCREENS_DIR is set, and
// saves one full-page PNG per screen, theme and width there, named
// <screen>.<theme>.<width>.png, so two builds of the console can be compared
// image by image (scripts/screens-diff.py). The first run on an install seeds
// a project with one image service through the API (deployed when a host is
// connected); later runs reuse it, so a copy of the same data directory gives
// the same screens to every build.
import { expect, test, type Page } from "@playwright/test";
import { mkdirSync } from "node:fs";
import { join } from "node:path";
import { apiGet, until } from "../../harness/api";
import { env } from "../../harness/env";

const dir = process.env.E2E_SCREENS_DIR || "";
const PROJECT = "screens";
const THEMES = ["dark", "light"] as const;
const WIDTHS = [1440, 1024, 390];
const INSTALL_SECTIONS = ["general", "hosts", "github", "registry-credentials", "backups", "members", "api-tokens", "upgrades", "usage", "audit-log"];

test.skip(!dir, "set E2E_SCREENS_DIR to capture screens");
test.describe.configure({ mode: "serial" });

interface Seed { pid: string; eid: string; sid: string; did?: string }
let seed: Seed;

const projectExists = async (page: Page) =>
  (await apiGet<{ name: string }[]>(page, "/projects")).some((x) => x.name === PROJECT);

async function findSeed(page: Page): Promise<Seed | undefined> {
  const projects = await apiGet<{ id: string; name: string; environments: { id: string }[] }[]>(page, "/projects");
  for (const p of projects.filter((x) => x.name === PROJECT)) {
    const eid = p.environments[0]?.id;
    if (!eid) continue;
    const cv = await apiGet<{ services: { id: string; config: { name: string }; activeDeployment?: string }[] }>(page, `/environments/${eid}/canvas`);
    const s = cv.services.find((x) => x.config.name === "whoami");
    if (!s) continue;
    // A console without a host never deploys, so the build log screen needs a deployment to exist.
    const deps = await apiGet<{ id: string }[]>(page, `/environments/${eid}/services/${s.id}/deployments`);
    return { pid: p.id, eid, sid: s.id, did: s.activeDeployment || deps[0]?.id };
  }
  return undefined;
}

test("seed the screens project", async ({ page }) => {
  // Reuse the project when it is there, so a copied data directory gives every build the same screens.
  if (await projectExists(page)) { seed = await until(() => findSeed(page), 30_000); return; }
  const post = async (path: string, data: unknown) => {
    const res = await page.request.post(`/api/v1${path}`, { data });
    expect(res.ok(), `POST ${path}: ${res.status()}`).toBe(true);
    return res.json();
  };
  const project = await post("/projects", { name: PROJECT });
  const eid = project.environments[0].id;
  await post(`/environments/${eid}/services`, { name: "whoami", source: { type: "image", image: env.whoamiImage } });
  await post(`/environments/${eid}/changes/commit`, { message: "Seed the screens project", deploy: true });
  seed = await until(() => findSeed(page), 120_000);
});

function screens(s: Seed): [string, string][] {
  const env = `?environment=${s.eid}`;
  return [
    ["projects", "/"],
    ["new-project", "/new"],
    ["templates", "/new/template"],
    ["canvas-service", `/project/${s.pid}/service/${s.sid}${env}`],
    ...(s.did ? [["deployment-logs", `/project/${s.pid}/service/${s.sid}/deployment/${s.did}/build${env}`] as [string, string]] : []),
    ["logs", `/project/${s.pid}/logs${env}`],
    ["observability", `/project/${s.pid}/observability${env}`],
    ["project-settings", `/project/${s.pid}/settings/general${env}`],
    ...INSTALL_SECTIONS.map((x): [string, string] => [`settings-${x}`, `/settings/${x}`]),
    ["account", "/account/profile"],
    ["cli-auth", "/cli-auth"],
    ["console-popout", `/console-popout?service=${s.sid}&environment=${s.eid}`],
  ];
}

async function settle(page: Page) {
  // Live views keep a stream open, so the network never goes fully idle.
  await page.waitForLoadState("networkidle", { timeout: 3000 }).catch(() => undefined);
  await page.evaluate(() => document.fonts.ready);
  await page.waitForTimeout(400);
}

for (const theme of THEMES) {
  for (const width of WIDTHS) {
    test(`screens ${theme} ${width}`, async ({ page }) => {
      test.setTimeout(300_000);
      mkdirSync(dir, { recursive: true });
      await page.addInitScript((t) => localStorage.setItem("5050.theme", t), theme);
      await page.emulateMedia({ reducedMotion: "reduce" });
      await page.setViewportSize({ width, height: 900 });
      for (const [name, path] of screens(seed)) {
        await page.goto(path);
        await settle(page);
        await page.screenshot({ path: join(dir, `${name}.${theme}.${width}.png`), fullPage: true, animations: "disabled", caret: "hide" });
      }
      await page.context().clearCookies();
      await page.goto("/login");
      await settle(page);
      await page.screenshot({ path: join(dir, `login.${theme}.${width}.png`), fullPage: true, animations: "disabled", caret: "hide" });
    });
  }
}
