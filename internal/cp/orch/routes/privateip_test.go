// SPDX-License-Identifier: AGPL-3.0-only

// LuxRail: tests for names that can only point at a private address.
package routes

import "testing"

func TestPrivateIPName(t *testing.T) {
	cases := map[string]string{
		"console.10-10-180-5.sslip.io": "10.10.180.5",  // check-public: allow
		"10-10-180-5.sslip.io":         "10.10.180.5",  // check-public: allow
		"app.192.168.1.20.nip.io":      "192.168.1.20", // check-public: allow
		"api.127-0-0-1.sslip.io":       "127.0.0.1",
		"console.172-16-0-9.SSLIP.IO.": "172.16.0.9", // check-public: allow
		"console.203-0-113-7.sslip.io": "",
		"app.203.0.113.7.nip.io":       "",
		"console.example.com":          "",
		"10-10-180-5.example.com":      "", // check-public: allow
		"sslip.io":                     "",
		"console.not-an-ip.sslip.io":   "",
	}
	for in, want := range cases {
		if got := PrivateIPName(in); got != want {
			t.Errorf("PrivateIPName(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestTLSModeForPrivateNamesAndIPs(t *testing.T) {
	cases := map[string]string{
		"console.10-10-180-5.sslip.io": "internal", // check-public: allow
		"10.10.180.5":                  "internal", // check-public: allow
		"203.0.113.7":                  "internal",
		"console.203-0-113-7.sslip.io": "acme",
		"app.example.com":              "acme",
		"api.internal":                 "internal",
	}
	for in, want := range cases {
		if got := TLSMode(in); got != want {
			t.Errorf("TLSMode(%q) = %q, want %q", in, got, want)
		}
	}
}
