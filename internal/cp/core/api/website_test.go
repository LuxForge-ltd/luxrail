// SPDX-License-Identifier: AGPL-3.0-only

// LuxRail: the website link of a service through the API.
package api

import (
	"testing"

	"github.com/luxforge-ltd/luxrail/pkg/api"
)

func TestServiceWebsiteLink(t *testing.T) {
	e := newEnv(t)
	owner := e.client()
	tok, err := e.c.CreateSetupToken(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	owner.do("POST", "/api/v1/setup", api.SetupRequest{Token: tok, Email: "owner@example.com", Name: "Owner", Password: "correct horse battery"}, nil, 200)
	var prj api.Project
	owner.do("POST", "/api/v1/projects", api.CreateProjectRequest{Name: "shop"}, &prj, 200)
	eid := prj.Environments[0].ID
	var cs api.CreateServiceResponse
	owner.do("POST", "/api/v1/environments/"+eid+"/services", api.CreateServiceRequest{Name: "web", DeployNow: true,
		Source: api.SourceConfig{Type: "image", Image: "nginx:1.27"}}, &cs, 200)
	path := "/api/v1/environments/" + eid + "/services/" + cs.Services[0].ID

	yes, no := true, false
	site, plain := "https://shop.example.com", "http://shop.example.com"
	owner.do("PATCH", path+"/info", api.UpdateServiceInfoRequest{Website: &site}, nil, 422)
	owner.do("PATCH", path+"/info", api.UpdateServiceInfoRequest{WebFrontend: &yes, Website: &plain}, nil, 422)
	var svc api.Service
	owner.do("PATCH", path+"/info", api.UpdateServiceInfoRequest{WebFrontend: &yes, Website: &site}, &svc, 200)
	owner.do("GET", path, nil, &svc, 200)
	if !svc.Config.WebFrontend || svc.Config.Website != site {
		t.Fatalf("website not stored: %+v", svc.Config)
	}
	var ch api.ChangesResponse
	owner.do("GET", "/api/v1/environments/"+eid+"/changes", nil, &ch, 200)
	if len(ch.Changes) != 0 {
		t.Fatalf("setting the website staged changes: %+v", ch.Changes)
	}
	owner.do("PATCH", path+"/config", api.UpdateServiceConfigRequest{Patch: map[string]any{"website": site}}, nil, 422)
	owner.do("PATCH", path+"/info", api.UpdateServiceInfoRequest{WebFrontend: &no}, nil, 200)
	var after api.Service
	owner.do("GET", path, nil, &after, 200)
	if after.Config.WebFrontend || after.Config.Website != "" {
		t.Fatalf("clearing the flag kept the website: %+v", after.Config)
	}
}
