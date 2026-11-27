// SPDX-License-Identifier: AGPL-3.0-only

// LuxRail: changing a member's email through the API.
package api

import (
	"strings"
	"testing"

	"github.com/luxforge-ltd/luxrail/pkg/api"
)

// join invites a user with the given role and signs them up.
func join(t *testing.T, e *env, owner *client, role api.Role, email string) (*client, api.User) {
	t.Helper()
	var inv api.Invite
	owner.do("POST", "/api/v1/invites", api.CreateInviteRequest{Role: role}, &inv, 200)
	c := e.client()
	var me api.Me
	c.do("POST", "/api/v1/invites/accept", api.AcceptInviteRequest{Token: inv.URL[strings.LastIndex(inv.URL, "/")+1:], Email: email, Name: string(role), Password: "another long password"}, &me, 200)
	return c, me.User
}

func TestMemberEmailChange(t *testing.T) {
	e := newEnv(t)
	owner := e.client()
	tok, err := e.c.CreateSetupToken(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	owner.do("POST", "/api/v1/setup", api.SetupRequest{Token: tok, Email: "owner@example.com", Name: "Owner", Password: "correct horse battery"}, nil, 200)
	var om api.Me
	owner.do("GET", "/api/v1/me", nil, &om, 200)
	ou := om.User
	admin, au := join(t, e, owner, api.RoleAdmin, "admin@example.com")
	member, mu := join(t, e, owner, api.RoleMember, "member@example.com")
	_, vu := join(t, e, owner, api.RoleViewer, "viewer@example.com")
	email := func(s string) api.UpdateMemberRequest { return api.UpdateMemberRequest{Email: &s} }

	var got api.User
	owner.do("PATCH", "/api/v1/members/"+vu.ID, email(" Reader@Example.com "), &got, 200)
	if got.Email != "reader@example.com" || got.EmailVerified {
		t.Fatalf("owner's change: %+v", got)
	}
	admin.do("PATCH", "/api/v1/members/"+mu.ID, email("dev@example.com"), &got, 200)
	if got.Email != "dev@example.com" {
		t.Fatalf("admin's change: %+v", got)
	}
	admin.do("PATCH", "/api/v1/members/"+au.ID, email("admin2@example.com"), nil, 200)
	owner.do("PATCH", "/api/v1/members/"+ou.ID, email("boss@example.com"), nil, 200)

	admin.do("PATCH", "/api/v1/members/"+ou.ID, email("taken-over@example.com"), nil, 403)
	member.do("PATCH", "/api/v1/members/"+vu.ID, email("x@example.com"), nil, 403)
	owner.do("PATCH", "/api/v1/members/"+vu.ID, email("DEV@example.com"), nil, 409)
	for _, bad := range []string{"", "not-an-email", "a@b@example.com", "Name <a@example.com>"} {
		owner.do("PATCH", "/api/v1/members/"+vu.ID, email(bad), nil, 422)
	}

	// The new address signs in; the old one does not.
	if code := login(t, e, "198.51.100.9", "dev@example.com", "another long password"); code != 200 {
		t.Fatalf("login with the new address: %d", code)
	}
	if code := login(t, e, "198.51.100.10", "member@example.com", "another long password"); code == 200 {
		t.Fatal("the old address still signs in")
	}
	var out api.Page[api.AuditEntry]
	owner.do("GET", "/api/v1/audit", nil, &out, 200)
	found := false
	for _, a := range out.Items {
		found = found || (a.Action == "member.update" && a.Detail == "email member@example.com to dev@example.com")
	}
	if !found {
		t.Fatalf("no audit entry with both addresses: %+v", out.Items)
	}
}
