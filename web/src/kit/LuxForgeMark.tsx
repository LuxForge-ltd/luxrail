// SPDX-License-Identifier: AGPL-3.0-only

// LuxRail: the LuxForge mark (from the LuxForge website's favicon and Logo), used as the
// console's logo. The colours are fixed so the mark reads the same in both themes; the hairline
// in currentColor keeps the ink tile visible on a dark page, as LuxRail's mark did.
import { PRODUCT_NAME } from "./brand";

/**
 * The LuxForge mark as an inline SVG.
 * @param size width and height in pixels (default 24)
 */
export function LuxForgeMark({ size = 24 }: { size?: number }) {
  return (
    <svg width={size} height={size} viewBox="0 0 32 32" fill="none" role="img" aria-label={PRODUCT_NAME}>
      <rect width="32" height="32" rx="8" fill="#0e0e12" />
      <rect x="0.5" y="0.5" width="31" height="31" rx="7.5" fill="none" stroke="currentColor" strokeOpacity="0.16" />
      <path d="M9 8.5L16 23L23 8.5H20.5L16 18.5L11.5 8.5H9Z" fill="#8b54f5" />
      <path d="M9 8.5L16 23L23 8.5" stroke="#ab7eff" strokeWidth="1" strokeOpacity="0.4" />
    </svg>
  );
}
