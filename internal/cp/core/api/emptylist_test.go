// SPDX-License-Identifier: AGPL-3.0-only

// LuxRail: tests that empty lists reach the console as [] and not null.
package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/luxforge-ltd/luxrail/pkg/api"
)

func TestEmptyListsAsArrays(t *testing.T) {
	var none []api.Invite
	for _, tc := range []struct {
		name string
		in   any
		want string
	}{
		{"nil slice", none, "[]"},
		{"empty slice", []string{}, "[]"},
		{"full slice", []string{"a"}, `["a"]`},
		{"struct", api.VersionInfo{Version: "v1"}, `{"version":"v1"`},
		{"nil", nil, "null"},
	} {
		b, err := json.Marshal(emptyListsAsArrays(tc.in))
		if err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		if !strings.HasPrefix(string(b), tc.want) {
			t.Errorf("%s: got %s, want %s", tc.name, b, tc.want)
		}
	}
}

// TestEmptyListPagesAnswerArrays covers the two console pages that hung: the
// Members page (no open invites) and the API tokens panel (no tokens).
func TestEmptyListPagesAnswerArrays(t *testing.T) {
	e := newEnv(t)
	owner := e.client()
	tok, err := e.c.CreateSetupToken(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	owner.do("POST", "/api/v1/setup", api.SetupRequest{Token: tok, Email: "owner@example.com", Name: "Owner", Password: "correct horse battery"}, nil, 200)
	for _, path := range []string{"/api/v1/invites", "/api/v1/tokens"} {
		res, err := owner.h.Get(e.srv.URL + path)
		if err != nil {
			t.Fatal(err)
		}
		b, _ := io.ReadAll(res.Body)
		_ = res.Body.Close()
		if res.StatusCode != http.StatusOK || strings.TrimSpace(string(b)) != "[]" {
			t.Errorf("GET %s: %d %s, want 200 []", path, res.StatusCode, b)
		}
	}
}
