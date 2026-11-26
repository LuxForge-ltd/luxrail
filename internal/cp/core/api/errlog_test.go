// SPDX-License-Identifier: AGPL-3.0-only

// LuxRail: tests that unexpected errors are logged with a reference.
package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/musakkka/openrail/internal/cp/core/fail"
	"github.com/musakkka/openrail/pkg/api"
)

func TestUnexpectedErrorsAreLoggedWithAReference(t *testing.T) {
	for _, tc := range []struct {
		name  string
		err   error
		cause string
	}{
		{"plain error", errors.New("disk I/O error"), "disk I/O error"},
		{"internal", fail.Internal(errors.New("open settings smtp: bad key")), "open settings smtp: bad key"},
	} {
		var logs bytes.Buffer
		s := &Server{Log: slog.New(slog.NewTextHandler(&logs, nil))}
		w := httptest.NewRecorder()
		s.writeErrLogged(w, httptest.NewRequest("GET", "/api/v1/tokens", nil), "GET /tokens", tc.err)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("%s: status %d", tc.name, w.Code)
		}
		var body struct {
			api.Error
			Details map[string]string `json:"details"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		ref := body.Details["ref"]
		if !strings.HasPrefix(ref, "e-") || !strings.Contains(body.Error.Error, ref) {
			t.Errorf("%s: answer %q has no reference", tc.name, w.Body.String())
		}
		line := logs.String()
		for _, want := range []string{"level=ERROR", "ref=" + ref, "route=\"GET /tokens\"", tc.cause} {
			if !strings.Contains(line, want) {
				t.Errorf("%s: log %q lacks %q", tc.name, line, want)
			}
		}
	}
}

func TestExpectedErrorsAreNotLogged(t *testing.T) {
	var logs bytes.Buffer
	s := &Server{Log: slog.New(slog.NewTextHandler(&logs, nil))}
	for _, err := range []error{fail.NotFound("No such token."), fail.Validation("name", "Name it."), fail.Sentence("Not now.")} {
		w := httptest.NewRecorder()
		s.writeErrLogged(w, httptest.NewRequest("GET", "/api/v1/tokens", nil), "GET /tokens", err)
		if w.Code >= 500 || strings.Contains(w.Body.String(), "ref") {
			t.Errorf("%v: answered %d %s", err, w.Code, w.Body.String())
		}
	}
	if logs.Len() != 0 {
		t.Errorf("expected errors were logged: %s", logs.String())
	}
}
