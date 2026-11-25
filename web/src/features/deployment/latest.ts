// SPDX-License-Identifier: AGPL-3.0-only

// "Deploy latest commit": deploys the head of a GitHub service's branch, or
// pulls a service's image again, from anywhere in the console (the
// Deployments tab, the canvas menu of a service and the command palette).

import type { SourceConfig } from "../../api/types.gen";

/** Services whose newest source can be deployed: GitHub repositories and images. */
export function canDeployLatest(source: Pick<SourceConfig, "type"> | undefined): boolean {
  return source?.type === "github" || source?.type === "image";
}

/** The action's label for a source. */
export function deployLatestLabel(source: Pick<SourceConfig, "type"> | undefined): string {
  return source?.type === "image" ? "Pull and redeploy image" : "Deploy latest commit";
}

/** The toast shown once the deployment has started. */
export function deployLatestStarted(source: Pick<SourceConfig, "type"> | undefined, name?: string): string {
  const what = source?.type === "image" ? "Pulling the newest image" : "Deploying the latest commit";
  return name ? `${what} of ${name}` : what;
}

/** The request body: `ref: "latest"` builds the head of the branch (or pulls the image again). */
export const DEPLOY_LATEST = { ref: "latest" } as const;
