// SPDX-License-Identifier: AGPL-3.0-only

// The same mark as kit/Logo, kept for older imports.
import { Logo } from "../kit/Logo";
import { PRODUCT_NAME } from "../kit/brand";

export default Logo;

export function Wordmark() {
  return <span className="text-[15px] font-semibold tracking-tight">{PRODUCT_NAME}</span>;
}
