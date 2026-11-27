// SPDX-License-Identifier: AGPL-3.0-only

// LuxRail: names and detection for the CLI downloads the server hands out.
import type * as T from "../../api/types.gen";

const OS_NAMES: Record<string, string> = { windows: "Windows", darwin: "macOS", linux: "Linux" };
const ARCH_NAMES: Record<string, Record<string, string>> = {
  darwin: { arm64: "Apple silicon", amd64: "Intel" },
  default: { amd64: "x64", arm64: "ARM64" },
};

/** "macOS (Apple silicon)", "Windows (x64)". */
export function platformLabel(d: Pick<T.CLIDownload, "os" | "arch">): string {
  const arch = (ARCH_NAMES[d.os] ?? ARCH_NAMES.default)[d.arch] ?? d.arch;
  return `${OS_NAMES[d.os] ?? d.os} (${arch})`;
}

/** The download that best fits the browser's system, from its user agent and platform strings. */
export function detectDownload(downloads: T.CLIDownload[], ua: string, platform = ""): T.CLIDownload | undefined {
  const s = `${ua} ${platform}`.toLowerCase();
  const os = /windows|win32|win64/.test(s) ? "windows" : /mac|darwin/.test(s) ? "darwin" : /linux|x11|cros/.test(s) ? "linux" : "";
  const arm = /arm64|aarch64|armv8/.test(s);
  const ofOS = downloads.filter((d) => d.os === os);
  // Browsers on Apple silicon still report "Intel", so prefer arm64 on macOS.
  const arch = os === "darwin" ? "arm64" : arm ? "arm64" : "amd64";
  return ofOS.find((d) => d.arch === arch) ?? ofOS[0] ?? downloads[0];
}

/** The shell steps after downloading file on os. */
export function installSteps(os: string, file: string): string[] {
  if (os === "windows") {
    return [
      `Expand-Archive -Path "$HOME\\Downloads\\${file}" -DestinationPath "$HOME\\bin" -Force`,
      `[Environment]::SetEnvironmentVariable("Path", [Environment]::GetEnvironmentVariable("Path", "User") + ";$HOME\\bin", "User")`,
    ];
  }
  const steps = [`mkdir -p ~/.local/bin && tar -xzf ~/Downloads/${file} -C ~/.local/bin luxrail`];
  if (os === "darwin") steps.push("xattr -d com.apple.quarantine ~/.local/bin/luxrail");
  return steps;
}
