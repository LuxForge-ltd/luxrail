// SPDX-License-Identifier: AGPL-3.0-only
import { describe, expect, it } from "vitest";
import { DEPLOY_LATEST, canDeployLatest, deployLatestLabel, deployLatestStarted } from "./latest";

describe("deploy latest", () => {
  it("is offered for GitHub and image services only", () => {
    expect(canDeployLatest({ type: "github" })).toBe(true);
    expect(canDeployLatest({ type: "image" })).toBe(true);
    expect(canDeployLatest({ type: "upload" })).toBe(false);
    expect(canDeployLatest({ type: "empty" })).toBe(false);
    expect(canDeployLatest(undefined)).toBe(false);
  });
  it("names the action after the source", () => {
    expect(deployLatestLabel({ type: "github" })).toBe("Deploy latest commit");
    expect(deployLatestLabel({ type: "image" })).toBe("Pull and redeploy image");
  });
  it("says what started, with the service name when given", () => {
    expect(deployLatestStarted({ type: "github" }, "web")).toBe("Deploying the latest commit of web");
    expect(deployLatestStarted({ type: "image" })).toBe("Pulling the newest image");
  });
  it("asks the API for the latest ref", () => expect(DEPLOY_LATEST).toEqual({ ref: "latest" }));
});
