// SPDX-License-Identifier: AGPL-3.0-only

// LuxRail: tests for the console route on a private console name's IP.
package engine

import "testing"

func TestConsoleIPRoute(t *testing.T) {
	r := consoleIPRoute("console.10-10-180-5.sslip.io", 5050) // check-public: allow

	if r == nil || r.Domain != "10.10.180.5" || r.TLS != "internal" || r.UpstreamName != consoleUpstream || r.Port != 5050 { // check-public: allow
		t.Fatalf("private sslip name: %+v", r)
	}
	for _, d := range []string{"console.203-0-113-7.sslip.io", "console.example.com", ""} {
		if r := consoleIPRoute(d, 5050); r != nil {
			t.Errorf("%q: got %+v, want none", d, r)
		}
	}
}
