// SPDX-License-Identifier: AGPL-3.0-only

// LuxRail: who may change a member's email in the console.
import type * as T from "../../api/types.gen";

// canChangeEmail says whether the signed-in person may change u's email: owners
// change anyone's, admins anyone's but the owner's. The API enforces the same.
export function canChangeEmail(me: Pick<T.User, "role">, u: Pick<T.User, "role">): boolean {
  if (me.role === "owner") return true;
  return me.role === "admin" && u.role !== "owner";
}
