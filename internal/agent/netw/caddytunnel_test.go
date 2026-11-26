// SPDX-License-Identifier: AGPL-3.0-only

// LuxRail: tests that HTTPS-forwarded requests on :80 are served, not redirected.
package netw

import (
	"encoding/json"
	"testing"
)

func TestForwardedHTTPSIsServedNotRedirected(t *testing.T) {
	b, errs := renderCaddy(testParams, testIngress())
	if len(errs) != 0 {
		t.Fatal(errs)
	}
	var cfg struct {
		Apps struct {
			HTTP struct {
				Servers map[string]struct {
					Routes []struct {
						Match []struct {
							Host   []string            `json:"host"`
							Header map[string][]string `json:"header"`
						} `json:"match"`
						Handle []map[string]any `json:"handle"`
					} `json:"routes"`
				} `json:"servers"`
			} `json:"http"`
		} `json:"apps"`
	}
	if err := json.Unmarshal(b, &cfg); err != nil {
		t.Fatal(err)
	}
	forwarded := map[string]int{}
	redirectAt := -1
	for i, r := range cfg.Apps.HTTP.Servers["http"].Routes {
		if len(r.Handle) == 1 && r.Handle[0]["status_code"] == float64(308) {
			redirectAt = i
		}
		if len(r.Match) == 1 && len(r.Match[0].Header["X-Forwarded-Proto"]) == 1 {
			if r.Handle[0]["handler"] != "reverse_proxy" {
				t.Errorf("forwarded route for %v is not a proxy: %v", r.Match[0].Host, r.Handle)
			}
			forwarded[r.Match[0].Host[0]] = i
		}
	}
	if redirectAt < 0 {
		t.Fatal("no 308 route: plain HTTP must still be redirected")
	}
	// Redirected domains (acme and internal) get a forwarded route ahead of
	// the 308; plain.example.com (tls off) is served on :80 anyway.
	for _, d := range []string{"console.example.com", "api.example.com", "web-ab12.203-0-113-5.sslip.io", "dev.internal.test"} {
		i, ok := forwarded[d]
		if !ok || i > redirectAt {
			t.Errorf("%s: forwarded route at %d (present %v), redirect at %d", d, i, ok, redirectAt)
		}
	}
	if _, ok := forwarded["plain.example.com"]; ok {
		t.Error("plain.example.com needs no forwarded route")
	}
}
