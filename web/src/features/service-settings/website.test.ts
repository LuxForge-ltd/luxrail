// SPDX-License-Identifier: AGPL-3.0-only

import { describe, expect, it } from "vitest";
import { websiteError } from "./website";

describe("websiteError", () => {
  it("accepts https addresses", () => {
    expect(websiteError("https://app.example.com")).toBeNull();
    expect(websiteError(" https://app.example.com/path?q=1 ")).toBeNull();
  });
  it("refuses anything else", () => {
    for (const v of ["http://app.example.com", "app.example.com", "javascript:alert(1)", "https://user:pw@app.example.com", "ftp://x.example.com"]) {
      expect(websiteError(v)).not.toBeNull();
    }
  });
});
