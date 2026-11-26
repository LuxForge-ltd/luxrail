// SPDX-License-Identifier: AGPL-3.0-only

// LuxRail: tests for the LAN console server in the Caddy config.
package netw

import (
	"context"
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"github.com/luxforge-ltd/luxrail/pkg/proto"
)

func TestConsoleLAN(t *testing.T) {
	in := testIngress()
	in.ConsoleLAN = &proto.ConsoleLAN{Listen: "192.0.2.5:8443", Target: "127.0.0.1:5050"}
	b, errs := renderCaddy(testParams, in)
	if len(errs) != 0 {
		t.Fatal(errs)
	}
	var cfg struct {
		Apps struct {
			HTTP struct {
				Servers map[string]struct {
					Listen []string         `json:"listen"`
					TLS    []map[string]any `json:"tls_connection_policies"`
					Routes []struct {
						Handle []struct {
							Handler   string `json:"handler"`
							Upstreams []struct {
								Dial string `json:"dial"`
							} `json:"upstreams"`
						} `json:"handle"`
					} `json:"routes"`
				} `json:"servers"`
			} `json:"http"`
			TLS struct {
				Certificates struct {
					Automate []string `json:"automate"`
				} `json:"certificates"`
				Automation struct {
					Policies []struct {
						Subjects []string `json:"subjects"`
						Issuers  []struct {
							Module string `json:"module"`
						} `json:"issuers"`
					} `json:"policies"`
				} `json:"automation"`
			} `json:"tls"`
		} `json:"apps"`
	}
	if err := json.Unmarshal(b, &cfg); err != nil {
		t.Fatal(err)
	}
	lan, ok := cfg.Apps.HTTP.Servers["console_lan"]
	if !ok {
		t.Fatal("no console_lan server")
	}
	if !slices.Equal(lan.Listen, []string{"192.0.2.5:8443"}) || len(lan.TLS) != 1 {
		t.Errorf("listen %v, tls policies %v", lan.Listen, lan.TLS)
	}
	if h := lan.Routes[0].Handle[0]; h.Handler != "reverse_proxy" || h.Upstreams[0].Dial != "127.0.0.1:5050" {
		t.Errorf("handler %+v", h)
	}
	if !slices.Contains(cfg.Apps.TLS.Certificates.Automate, "192.0.2.5") {
		t.Errorf("the IP is not in automate: %v", cfg.Apps.TLS.Certificates.Automate)
	}
	internal := false
	for _, p := range cfg.Apps.TLS.Automation.Policies {
		if slices.Contains(p.Subjects, "192.0.2.5") && p.Issuers[0].Module == "internal" {
			internal = true
		}
	}
	if !internal {
		t.Error("the IP's certificate is not from the internal CA")
	}

	// Without the setting the config is exactly LuxRail's.
	plain, _ := renderCaddy(testParams, testIngress())
	in.ConsoleLAN = nil
	again, _ := renderCaddy(testParams, in)
	if string(plain) != string(again) {
		t.Error("no LAN console must render the same config as before")
	}

	in.ConsoleLAN = &proto.ConsoleLAN{Listen: "lan:8443", Target: "127.0.0.1:5050"}
	if _, errs := renderCaddy(testParams, in); len(errs) != 1 {
		t.Errorf("a bad address must be reported once, got %v", errs)
	}
}

// Caddy refuses a whole config when a listener cannot bind, so a LAN
// address that is not on this host must not take the routes down with it.
func TestConsoleLANNeedsTheHostAddress(t *testing.T) {
	g, fc, _, _ := newTestIngress(t)
	up := false
	g.hasAddr = func(string) bool { return up }
	in := testIngress()
	in.ConsoleLAN = &proto.ConsoleLAN{Listen: "192.0.2.5:8443", Target: "127.0.0.1:5050"}
	ctx := context.Background()
	if err := g.Apply(ctx, in, nil); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(fc.config), "console_lan") || fc.loads != 1 {
		t.Errorf("loads %d; console_lan present before its address exists", fc.loads)
	}
	up = true
	if err := g.Apply(ctx, in, nil); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(fc.config), "console_lan") {
		t.Error("console_lan was not added once the address exists")
	}
}
