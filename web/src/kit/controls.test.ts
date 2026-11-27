// SPDX-License-Identifier: AGPL-3.0-only
import { describe, expect, it } from "vitest";
import { fieldFills } from "./controls";

describe("fieldFills", () => {
  it("fills the row for a field select with no width of its own", () => {
    expect(fieldFills("field")).toBe(true);
    expect(fieldFills("field", "h-9")).toBe(true);
    expect(fieldFills("field", "min-w-0")).toBe(true);
  });
  it("keeps the caller's width", () => {
    expect(fieldFills("field", "w-28 h-9")).toBe(false);
    expect(fieldFills("field", "h-9 w-32")).toBe(false);
  });
  it("never fills for other variants", () => {
    expect(fieldFills("outline")).toBe(false);
    expect(fieldFills("flat", "w-auto")).toBe(false);
  });
});
