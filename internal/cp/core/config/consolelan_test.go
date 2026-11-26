// SPDX-License-Identifier: AGPL-3.0-only

// LuxRail: tests for the LAN console address in config.json.
package config

import "testing"

func TestValidateConsoleLAN(t *testing.T) {
	for v, ok := range map[string]bool{
		"":                   true,
		"10.10.180.5:8443":   true,
		"192.0.2.7:9443":     true,
		"10.10.180.5":        false,
		"console.lan:8443":   false,
		"0.0.0.0:8443":       false,
		"127.0.0.1:8443":     false,
		"[2001:db8::1]:8443": false,
		"10.10.180.5:443":    false,
		"10.10.180.5:5050":   false,
		"10.10.180.5:70000":  false,
	} {
		if err := ValidateConsoleLAN(v); (err == nil) != ok {
			t.Errorf("%q: %v, want ok=%v", v, err, ok)
		}
	}
}
