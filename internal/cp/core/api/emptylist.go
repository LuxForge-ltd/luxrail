// SPDX-License-Identifier: AGPL-3.0-only

// LuxRail: empty lists are answered as [] rather than null.
//
// Purpose: a handler that returns a nil slice (store.All over no rows, for
// example) would otherwise be encoded as JSON null. The console treats null
// as "not loaded yet", so pages such as Members (no open invites) and API
// tokens (no tokens) stayed on their loading state for ever.
// Scope: the top-level value of a JSON answer of the console API only.
// Operational role: called by the route wrapper in server.go just before the
// answer is written; it never changes a non-empty value.
package api

import "reflect"

// emptyListsAsArrays returns v unchanged unless it is a nil slice, which it
// replaces with an empty slice of the same type, so the answer encodes as [].
//
// Parameters: v is a handler's result. Returns the value to encode.
func emptyListsAsArrays(v any) any {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Slice || !rv.IsNil() {
		return v
	}
	return reflect.MakeSlice(rv.Type(), 0, 0).Interface()
}
