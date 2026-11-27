// SPDX-License-Identifier: AGPL-3.0-only
import { describe, expect, it } from "vitest";
import { detectDownload, installSteps, platformLabel } from "./platform";

const downloads = [
  { os: "windows", arch: "amd64", file: "luxrail-cli_1.0.11_windows_amd64.zip", path: "/api/v1/cli/download/windows/amd64" },
  { os: "darwin", arch: "arm64", file: "luxrail-cli_1.0.11_darwin_arm64.tar.gz", path: "/api/v1/cli/download/darwin/arm64" },
  { os: "darwin", arch: "amd64", file: "luxrail-cli_1.0.11_darwin_amd64.tar.gz", path: "/api/v1/cli/download/darwin/amd64" },
  { os: "linux", arch: "amd64", file: "luxrail-cli_1.0.11_linux_amd64.tar.gz", path: "/api/v1/cli/download/linux/amd64" },
  { os: "linux", arch: "arm64", file: "luxrail-cli_1.0.11_linux_arm64.tar.gz", path: "/api/v1/cli/download/linux/arm64" },
];

describe("CLI platforms", () => {
  it("labels each download", () => {
    expect(downloads.map(platformLabel)).toEqual(["Windows (x64)", "macOS (Apple silicon)", "macOS (Intel)", "Linux (x64)", "Linux (ARM64)"]);
  });
  it("picks the download for the browser's system", () => {
    expect(detectDownload(downloads, "Mozilla/5.0 (Windows NT 10.0; Win64; x64)", "Win32")?.os).toBe("windows");
    expect(detectDownload(downloads, "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)", "MacIntel")?.arch).toBe("arm64");
    expect(detectDownload(downloads, "Mozilla/5.0 (X11; Linux x86_64)", "Linux x86_64")?.arch).toBe("amd64");
    expect(detectDownload(downloads, "Mozilla/5.0 (X11; Linux aarch64)", "Linux aarch64")?.arch).toBe("arm64");
    expect(detectDownload(downloads, "something else")).toBe(downloads[0]);
    expect(detectDownload([], "Windows")).toBeUndefined();
  });
  it("gives install steps per system", () => {
    expect(installSteps("windows", "a.zip")[0]).toContain("Expand-Archive");
    expect(installSteps("darwin", "a.tar.gz")).toHaveLength(2);
    expect(installSteps("linux", "a.tar.gz")[0]).toContain("tar -xzf ~/Downloads/a.tar.gz -C ~/.local/bin luxrail");
  });
});
