// SPDX-License-Identifier: AGPL-3.0-only

// LuxRail: the product name in tab titles.
import { describe, expect, it } from "vitest";
import { PRODUCT_NAME, tabTitle } from "./brand";

describe("tabTitle", () => {
  it("names the page and the product", () => {
    expect(tabTitle("Projects")).toBe("Projects | LuxRail");
  });
  it("is the product alone without a page", () => {
    expect(tabTitle("")).toBe(PRODUCT_NAME);
  });
});
