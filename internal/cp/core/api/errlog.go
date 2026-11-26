// SPDX-License-Identifier: AGPL-3.0-only

// LuxRail: unexpected errors are logged before they are answered.
//
// Purpose: an error the API does not recognise is answered 500 "Something
// went wrong on the server. The details are in the 5050d log." openrail
// v0.1.9 wrote that answer without logging anything, so the log held no
// details. This file logs the cause with the route and a short reference,
// and puts the same reference in the answer so the two can be matched.
// Scope: errors returned by API route handlers and by authentication.
// Operational role: used by the route wrapper in server.go in place of
// writeErr; expected errors (validation, not found, ...) are answered as
// before and not logged.
package api

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/musakkka/openrail/internal/cp/core/fail"
	"github.com/musakkka/openrail/pkg/api"
)

// writeErrLogged answers err like writeErr, and when the answer is an
// internal error it first logs the cause at ERROR with the method, the route
// pattern, the path and a reference that the answer repeats.
//
// Parameters: w and r are the request's writer and request; route is the
// route pattern (for example "GET /tokens"); err is the handler's error.
func (s *Server) writeErrLogged(w http.ResponseWriter, r *http.Request, route string, err error) {
	code, body := fail.ToAPI(err)
	if body.Code == api.CodeInternal {
		ref := errorRef()
		s.Log.Error("unexpected error answered 500", "ref", ref, "method", r.Method, "route", route,
			"path", r.URL.Path, "err", errorCause(err))
		body.Error += " Reference " + ref + "."
		body.Details = map[string]string{"ref": ref}
	}
	writeJSON(w, code, body)
}

// errorCause returns the error worth logging: the hidden cause of a
// fail.Internal error (whose own text is the generic sentence), else err.
//
// Parameters: err is a handler's error. Returns the cause, never nil for a
// non-nil err.
func errorCause(err error) error {
	if e := fail.As(err); e != nil && e.Err != nil {
		return e.Err
	}
	return err
}

// errorRef returns a short random reference such as "e-3f9a12bc", which
// ties an answer to its log line without exposing anything about the error.
func errorRef() string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return "e-" + hex.EncodeToString(b)
}
