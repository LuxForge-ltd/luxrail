// SPDX-License-Identifier: AGPL-3.0-only

// The console's mark. LuxRail shows the LuxForge mark (kit/LuxForgeMark.tsx) instead of
// openrail's Turnout mark; every caller keeps importing Logo from here.
import { LuxForgeMark } from "./LuxForgeMark";

export function Logo({ size = 24 }: { size?: number }) {
  return <LuxForgeMark size={size} />;
}
