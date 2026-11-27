// SPDX-License-Identifier: AGPL-3.0-only

import { describe, expect, it } from "vitest";
import { canChangeEmail } from "./email";

describe("canChangeEmail", () => {
  it("lets the owner change anyone's email", () => {
    for (const role of ["owner", "admin", "member", "viewer"]) expect(canChangeEmail({ role: "owner" }, { role })).toBe(true);
  });
  it("lets admins change everyone's but the owner's", () => {
    expect(canChangeEmail({ role: "admin" }, { role: "owner" })).toBe(false);
    for (const role of ["admin", "member", "viewer"]) expect(canChangeEmail({ role: "admin" }, { role })).toBe(true);
  });
  it("never lets members or viewers", () => {
    for (const me of ["member", "viewer"]) expect(canChangeEmail({ role: me }, { role: "viewer" })).toBe(false);
  });
});
