// SPDX-License-Identifier: AGPL-3.0-only

// LuxRail: the check for a service website address, matching the API's.

/** websiteError returns why v is not an https address, or null. */
export function websiteError(v: string): string | null {
  try {
    const u = new URL(v.trim());
    if (u.protocol !== "https:" || !u.hostname || u.username || u.password) throw new Error();
    return null;
  } catch {
    return "Use an https:// address, such as https://app.example.com.";
  }
}
