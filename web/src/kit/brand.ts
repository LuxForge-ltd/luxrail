// SPDX-License-Identifier: AGPL-3.0-only

// LuxRail: the product name the console shows (tab title, header, logo label).
// Kept in one place so upstream text changes and LuxForge branding stay apart.
export const PRODUCT_NAME = "LuxRail";

/**
 * Builds a browser tab title such as "Projects | LuxRail".
 * @param page the page name, or "" for the product name alone
 * @returns the title text
 */
export function tabTitle(page: string): string {
  return page ? `${page} | ${PRODUCT_NAME}` : PRODUCT_NAME;
}
